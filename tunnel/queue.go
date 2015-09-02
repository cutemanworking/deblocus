package tunnel

import (
	"container/list"
	ex "github.com/spance/deblocus/exception"
	log "github.com/spance/deblocus/golang/glog"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	TCP_CLOSE_R uint8 = 1
	TCP_CLOSE_W uint8 = 1 << 1
	TCP_CLOSED        = TCP_CLOSE_R | TCP_CLOSE_W
)
const (
	// close code
	CLOSED_FORCE = iota
	CLOSED_WRITE
	CLOSED_BY_ERR
)

const (
	TICKER_INTERVAL = time.Second * 15
)

type edgeConn struct {
	mux      *multiplexer
	tun      *Conn
	conn     net.Conn
	ready    chan byte // peer status
	key      string
	dest     string
	queue    *equeue
	positive bool // positively open
	closed   uint8
}

func newEdgeConn(mux *multiplexer, key, dest string, tun *Conn, conn net.Conn) *edgeConn {
	var edge = &edgeConn{
		mux:  mux,
		tun:  tun,
		conn: conn,
		key:  key,
		dest: dest,
	}
	if mux.isClient {
		edge.ready = make(chan byte, 1)
	}
	return edge
}

func (e *edgeConn) deliver(frm *frame) {
	if e.queue != nil {
		frm.conn = e
		e.queue._push(frm)
	}
}

// ------------------------------
// EgressRouter
// ------------------------------
type egressRouter struct {
	lock            *sync.RWMutex
	mux             *multiplexer
	registry        map[string]*edgeConn
	preRegistry     map[string]*list.List
	cleanerTicker   *time.Ticker
	stopCleanerChan chan bool
}

func newEgressRouter(mux *multiplexer) *egressRouter {
	r := &egressRouter{
		lock:            new(sync.RWMutex),
		mux:             mux,
		registry:        make(map[string]*edgeConn),
		cleanerTicker:   time.NewTicker(TICKER_INTERVAL),
		stopCleanerChan: make(chan bool, 1),
	}
	if !mux.isClient {
		r.preRegistry = make(map[string]*list.List)
	}
	go r.cleanTask()
	return r
}

func (r *egressRouter) preRegister(key string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.preRegistry[key] = list.New()
}

func (r *egressRouter) preDeliver(key string, f *frame) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if buffer := r.preRegistry[key]; buffer != nil {
		buffer.PushBack(f)
	}
}

func (r *egressRouter) removePreRegistered(key string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.preRegistry, key)
}

func (r *egressRouter) getRegistered(key string) (e *edgeConn, preRegistered bool) {
	r.lock.RLock()
	e = r.registry[key]
	_, preRegistered = r.preRegistry[key]
	r.lock.RUnlock()
	if e != nil && e.closed >= TCP_CLOSED {
		// clean when getting
		r.lock.Lock()
		delete(r.registry, key)
		r.lock.Unlock()
		return nil, false
	}
	return
}

func (r *egressRouter) clean() {
	defer func() {
		ex.CatchException(recover())
	}()
	r.lock.Lock()
	defer r.lock.Unlock()
	for k, e := range r.registry {
		// call conn.LocalAddr will give rise to checking fd.
		if e == nil || e.closed >= TCP_CLOSED || e.conn.LocalAddr() == nil {
			delete(r.registry, k)
		}
	}
}

func (r *egressRouter) register(key, destination string, tun *Conn, conn net.Conn, positive bool) *edgeConn {
	r.lock.Lock()
	defer r.lock.Unlock()
	var edge = r.registry[key]
	if edge == nil {
		edge = newEdgeConn(r.mux, key, destination, tun, conn)
		edge.positive = positive
		edge.initEqueue()
		r.registry[key] = edge
	}
	if buffer := r.preRegistry[key]; buffer != nil {
		delete(r.preRegistry, key)
		edge.queue._push_all(buffer)
	}
	return edge
}

// destroy whole router
func (r *egressRouter) destroy() {
	r.lock.Lock()
	defer r.lock.Unlock()
	var frm = &frame{action: FRAME_ACTION_CLOSE}
	for _, e := range r.registry {
		if e.queue != nil {
			e.queue._push(frm) // wakeup and self-exiting
		}
	}
	r.stopCleanTask()
	r.registry = nil
}

// remove edges (with queues) were related to the tun
func (r *egressRouter) cleanOfTun(tun *Conn) {
	r.lock.Lock()
	defer r.lock.Unlock()
	var prefix = tun.identifier
	var frm = &frame{action: FRAME_ACTION_CLOSE}
	for k, e := range r.registry {
		if strings.HasPrefix(k, prefix) && e.queue != nil {
			e.queue._push(frm)
			delete(r.registry, k)
		}
	}
}

func (r *egressRouter) cleanTask() {
	var (
		stopCh <-chan bool = r.stopCleanerChan
		runCh              = r.cleanerTicker.C
	)
	for {
		select {
		case s := <-stopCh:
			if s {
				return
			}
		case <-runCh:
			r.clean()
		}
	}
}

func (r *egressRouter) stopCleanTask() {
	r.stopCleanerChan <- true
	close(r.stopCleanerChan)
	r.cleanerTicker.Stop()
}

// -------------------------------
// Equeue
// -------------------------------
type equeue struct {
	edge   *edgeConn
	lock   sync.Locker
	cond   *sync.Cond
	buffer *list.List
}

func (edge *edgeConn) initEqueue() *equeue {
	l := new(sync.Mutex)
	q := &equeue{
		edge:   edge,
		lock:   l,
		cond:   sync.NewCond(l),
		buffer: list.New(),
	}
	edge.queue = q
	go q.sendLoop()
	return q
}

func (q *equeue) _push(frm *frame) {
	q.lock.Lock()
	defer q.cond.Signal()
	defer q.lock.Unlock()
	// push
	if q.buffer != nil {
		q.buffer.PushBack(frm)
	} // else the queue was exited
}

func (q *equeue) _push_all(buffer *list.List) {
	q.lock.Lock()
	defer q.cond.Signal()
	defer q.lock.Unlock()
	// push
	if _list := q.buffer; _list != nil {
		for i, e := buffer.Len(), buffer.Front(); i > 0; i, e = i-1, e.Next() {
			f := e.Value.(*frame)
			f.conn = q.edge
			_list.PushBack(f)
		}
	} // else the queue was exited
}

func (q *equeue) sendLoop() {
	for {
		var buffer *list.List
		q.lock.Lock()
		for q.buffer.Len() <= 0 {
			q.cond.Wait()
		}
		buffer = q.buffer
		q.buffer = list.New()
		q.lock.Unlock()

		for item := buffer.Front(); item != nil; item = item.Next() {
			// send
			var frm *frame = item.Value.(*frame)
			switch frm.action {
			case FRAME_ACTION_CLOSE:
				q._close(true, CLOSED_FORCE)
				return
			case FRAME_ACTION_CLOSE_W:
				q._close(false, CLOSED_WRITE)
				return
			default:
				werr := sendFrame(frm)
				if werr {
					edge := q.edge
					if edge.closed&TCP_CLOSE_W == 0 { // only positively closed can notify peer
						edge.closed |= TCP_CLOSE_W
						tun := edge.tun
						// may be a broken tun
						if tun == nil || tun.LocalAddr() == nil {
							tun = edge.mux.pool.Select()
						}
						if tun != nil {
							frm.length = 0
							frm.action = FRAME_ACTION_CLOSE_R
							tunWrite2(tun, frm)
						}
					}
					q._close(true, CLOSED_BY_ERR)
					return
				}
			}
		}
	}
}

// close for ending of queued task
func (q *equeue) _close(force bool, close_code uint) {
	q.lock.Lock()
	defer q.lock.Unlock()
	e := q.edge
	if log.V(4) {
		switch close_code {
		case CLOSED_BY_ERR:
			log.Infoln("terminate", e.dest)
		case CLOSED_FORCE:
			log.Infoln("close", e.dest)
		case CLOSED_WRITE:
			log.Infof("closeW %s by peer\n", e.dest)
		}
	}
	q.buffer.Init()
	q.buffer = nil
	if force {
		e.closed = TCP_CLOSE_R | TCP_CLOSE_W
		SafeClose(e.conn)
	} else {
		closeW(e.conn)
	}
}

func sendFrame(frm *frame) (werr bool) {
	dst := frm.conn.conn
	if log.V(5) {
		log.Infoln("SEND queue", frm)
	}
	dst.SetWriteDeadline(time.Now().Add(GENERAL_SO_TIMEOUT))
	nw, ew := dst.Write(frm.data)
	if nw == int(frm.length) && ew == nil {
		return
	}
	werr = true
	// an error occured
	log.Warningf("Write edge(%s) error(%v). %s\n", frm.conn.dest, ew, frm)
	return
}
package mempool

import (
	"container/list"
	"time"
)

var (
	Makes int
	Frees int
)

var (
	DefaultTimeout = time.Minute * 10
)

func makeBuffer() []byte {
	Makes += 1
	print("makes: ")
	print(Makes)
	print(", frees: ")
	print(Frees)
	print(", current: ")
	println(Makes - Frees)
	return make([]byte, 1024*1024*5)
}

type queued struct {
	when  time.Time
	slice []byte
}

func MakeRecycler() (get, give chan []byte) {
	get = make(chan []byte)
	give = make(chan []byte)

	go func() {
		q := new(list.List)
		for {
			if q.Len() == 0 {
				q.PushFront(queued{when: time.Now(), slice: makeBuffer()})
			}

			e := q.Front()

			timeout := time.NewTimer(time.Second * 3)
			select {
			case b := <-give:
				timeout.Stop()
				q.PushFront(queued{when: time.Now(), slice: b})

			case get <- e.Value.(queued).slice:
				timeout.Stop()
				q.Remove(e)

			case <-timeout.C:
				e := q.Front()
				for e != nil {
					n := e.Next()
					if time.Since(e.Value.(queued).when) > DefaultTimeout {
						Frees += 1
						q.Remove(e)
						e.Value = nil
					}
					e = n
				}
			}
		}

	}()

	return
}

var (
	GetBuffer  chan []byte
	GiveBuffer chan []byte
)

func init() {
	GetBuffer, GiveBuffer = MakeRecycler()
}

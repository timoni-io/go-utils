package set

import (
	"github.com/timoni-io/go-utils"
	"github.com/timoni-io/go-utils/slice"

	"golang.org/x/exp/constraints"
)

type Rigid[T comparable, S constraints.Unsigned] struct {
	slice.Rigid[T, S]
	set  Set[T]
	lock *utils.Lock
}

func NewRigid[T comparable, S constraints.Unsigned](size S) *Rigid[T, S] {
	return &Rigid[T, S]{
		Rigid: *slice.NewRigid[T](size),
		set:   *New[T](),
	}
}

func NewSafeRigid[T comparable, S constraints.Unsigned](size S) *Rigid[T, S] {
	return &Rigid[T, S]{
		Rigid: *slice.NewRigid[T](size),
		set:   *New[T](),
		lock:  &utils.Lock{},
	}
}

func (r *Rigid[T, S]) Add(x ...T) {
	toAdd := make([]T, 0, len(x))

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, v := range x {
		if !r.set.Contains(v) {
			toAdd = append(toAdd, v)
		}
	}

	r.set.Remove(r.Rigid.Add(toAdd...)...)
}

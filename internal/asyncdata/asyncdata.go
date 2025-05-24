package asyncdata

// AsyncData is a data structure to represent 4 states:
// - Initial
// - Loading
// - Failure
// - Success
//
// Originally written in TypeScript https://github.com/sectore/fees/blob/1ea522941ac2cf4a4310b018671d49fc5e1ae3e4/src/util/async.ts
type AsyncData[E error, A any] struct {
	state state
}

type state interface {
	isAsyncDataState()
}

type notAsked struct{}

func (notAsked) isAsyncDataState() {}

func NewNotAsked[E error, A any]() AsyncData[E, A] {
	return AsyncData[E, A]{state: notAsked{}}
}

func NotAsked[E error, A any](ad AsyncData[E, A]) bool {
	_, ok := ad.state.(notAsked)
	return ok
}

type loading[A any] struct {
	PrevData *A
}

func (loading[A]) isAsyncDataState() {}

func NewLoading[E error, A any](prevData *A) AsyncData[E, A] {
	return AsyncData[E, A]{state: loading[A]{PrevData: prevData}}
}

func Loading[E error, A any](ad AsyncData[E, A]) (*A, bool, bool) {
	if value, ok := ad.state.(loading[A]); ok {
		if value.PrevData != nil {
			return value.PrevData, true, true
		}
		return nil, false, true
	}
	return nil, false, false
}

type failure[E error] struct {
	Error E
}

func (failure[E]) isAsyncDataState() {}

func NewFailure[E error, A any](err E) AsyncData[E, A] {
	return AsyncData[E, A]{state: failure[E]{Error: err}}
}

func Failure[E error, A any](ad AsyncData[E, A]) (*E, bool) {
	if err, ok := ad.state.(failure[E]); ok {
		return &err.Error, true
	}
	return nil, false
}

type success[A any] struct {
	Data A
}

func (success[A]) isAsyncDataState() {}

func NewSuccess[E error, A any](data A) AsyncData[E, A] {
	return AsyncData[E, A]{state: success[A]{Data: data}}
}

func Success[E error, A any](ad AsyncData[E, A]) (*A, bool) {
	if succ, ok := ad.state.(success[A]); ok {
		return &succ.Data, true
	}
	return nil, false

}

func Map[E error, A any, B any](ad AsyncData[E, A], f func(A) B) AsyncData[E, B] {
	switch s := ad.state.(type) {
	case notAsked:
		return AsyncData[E, B]{state: notAsked{}}

	case loading[A]:
		if s.PrevData != nil {
			mapped := f(*s.PrevData)
			return AsyncData[E, B]{state: loading[B]{PrevData: &mapped}}
		}
		return AsyncData[E, B]{state: loading[B]{PrevData: nil}}

	case failure[E]:
		return AsyncData[E, B]{state: failure[E]{Error: s.Error}}

	case success[A]:
		return AsyncData[E, B]{state: success[B]{Data: f(s.Data)}}

	default:
		panic("unknown AsyncData state")
	}
}

func FoldA[E error, A any, T any](
	ad AsyncData[E, A],
	onNotAsked func() T,
	onLoading func(*A) T,
	onFailure func(E) T,
	onSuccess func(A) T,
) T {
	switch s := ad.state.(type) {
	case notAsked:
		return onNotAsked()
	case loading[A]:
		return onLoading(s.PrevData)
	case failure[E]:
		return onFailure(s.Error)
	case success[A]:
		return onSuccess(s.Data)
	default:
		panic("unknown AsyncData state")
	}

}

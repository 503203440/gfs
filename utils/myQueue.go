package utils

// Go 中，字段的导出（export）和非导出（unexported）是根据字段名的首字母大小写来区分的：
// 首字母大写的字段（如 Queue）是导出的，可以在包外访问。
// 首字母小写的字段（如 queue）是非导出的，只能在当前包内访问。
// 这里的Size是导出字段, queue是非导出字段

// 创建一个固定大小的队列,超过大小则会将最早加入的移除
type MyQueue struct {
	queue []any
	Size  int
}

// 入队方法
func (q *MyQueue) Enqueue(v any) {
	if len(q.queue) >= q.Size {
		// 在添加之前已经满了, 则移除第一个
		q.dequeue()
	}
	q.queue = append(q.queue, v)
}

// 出队方法
func (q *MyQueue) dequeue() (any, bool) {
	if len(q.queue) == 0 {
		return 0, false
	}
	v := (q.queue)[0]
	q.queue = (q.queue)[1:]
	return v, true
}

// 返回队列中的值
func (q *MyQueue) List() []any {
	return q.queue
}

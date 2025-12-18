package slice

// FilterMap 执行过滤并转化
// 如果 m 的第二个返回值是 false，那么我们会忽略第一个返回值
// 即便第二个返回值是 false，后续的元素也会被处理
func FilterMap[Src any, Dst any](src []Src, m func(idx int, src Src) (Dst, bool)) []Dst {
	res := make([]Dst, 0, len(src))
	for i, s := range src {
		dst, ok := m(i, s)
		if ok {
			res = append(res, dst)
		}
	}
	return res
}

func Map[Src any, Dst any](src []Src, m func(idx int, src Src) Dst) []Dst {
	dst := make([]Dst, 0, len(src))
	for i, s := range src {
		dst = append(dst, m(i, s))
	}
	return dst
}

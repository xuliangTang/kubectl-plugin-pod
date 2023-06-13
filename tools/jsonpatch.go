package tools

type JSONPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
type JSONPatchList []*JSONPatch

func AddJsonPatch(jps ...*JSONPatch) JSONPatchList {
	list := make([]*JSONPatch, len(jps))
	for index, jp := range jps {
		list[index] = jp
	}
	return list
}

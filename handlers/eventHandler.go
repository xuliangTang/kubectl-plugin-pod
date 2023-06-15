package handlers

type EventHandler struct{}

func (p EventHandler) OnAdd(obj interface{}) {
}

func (p EventHandler) OnUpdate(oldObj, newObj interface{}) {
}

func (p EventHandler) OnDelete(obj interface{}) {
}

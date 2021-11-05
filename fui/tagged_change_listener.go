package fui

type listenForChangeWithTag interface {
	onChange(tag string)
}

type taggedChangeListener struct {
	tag     string
	forward listenForChangeWithTag
}

func newTaggedChangeListener(tag string, listener listenForChangeWithTag) *taggedChangeListener {
	return &taggedChangeListener{tag, listener}
}

func (t *taggedChangeListener) DataChanged() {
	t.forward.onChange(t.tag)
}

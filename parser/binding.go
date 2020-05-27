package parser

type Binding struct {
	Parent *Binding
	NameValues map[string]*TypeTreeItem
}

func newBinding(parent *Binding) *Binding {
	return &Binding{
		Parent:     parent,
		NameValues: make(map[string]*TypeTreeItem),
	}
}

func (b *Binding) get(name string) *TypeTreeItem {
	v, ok := b.NameValues[name]
	if ok {
		return v
	}
	if b.Parent == nil {
		return nil
	}
	return b.Parent.get(name)
}

func (b *Binding) getOrElse(name string, orElse *TypeTreeItem) *TypeTreeItem {
	v := b.get(name)
	if v == nil {
		return orElse
	}
	return v
}

func (b *Binding) bind(name string, value *TypeTreeItem) {
	b.NameValues[name] = value
}
package hclwrite

type ObjectConsExpr struct {
	inTree
}

func newObjectConsExpr() *ObjectConsExpr {
	return &ObjectConsExpr{
		inTree: newInTree(),
	}
}

func (o *ObjectConsExpr) ValueFor(key string) *ObjectConsValue {
	var found *ObjectConsValue
	o.walkChildNodes(func(n *node) {
		if item, ok := n.content.(*ObjectConsItem); ok {
			k := item.key.content.(*ObjectConsKey)
			name := k.name.content.(*identifier)
			if k.literal && string(name.token.Bytes) == key {
				found = item.value.content.(*ObjectConsValue)
				return
			}
		}
	})
	return found
}

type ObjectConsItem struct {
	inTree
	key   *node
	value *node
}

func newObjectConsItem() *ObjectConsItem {
	item := &ObjectConsItem{
		inTree: newInTree(),
		key:    newNode(newObjectConsKey()),
		value:  newNode(newObjectConsValue()),
	}
	item.children.AppendNode(item.key)
	item.children.AppendNode(item.value)

	return item
}

func (item *ObjectConsItem) kv() (*ObjectConsKey, *ObjectConsValue) {
	return item.key.content.(*ObjectConsKey), item.value.content.(*ObjectConsValue)
}

type ObjectConsKey struct {
	inTree

	literal bool
	name    *node
}

func newObjectConsKey() *ObjectConsKey {
	return &ObjectConsKey{
		inTree: newInTree(),
	}
}

type ObjectConsValue struct {
	inTree
	expr *node
}

func newObjectConsValue() *ObjectConsValue {
	return &ObjectConsValue{
		inTree: newInTree(),
	}
}

func (o *ObjectConsValue) Expr() *Expression {
	return o.expr.content.(*Expression)
}

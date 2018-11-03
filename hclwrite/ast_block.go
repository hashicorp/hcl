package hclwrite

type Block struct {
	inTree

	leadComments *node
	typeName     *node
	labels       nodeSet
	open         *node
	body         *node
	close        *node
}

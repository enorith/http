package content

type FastHttpFileServer struct {
	*Response
	root         string
	stripSlashes int
}

func (ffs *FastHttpFileServer) Root() string {
	return ffs.root
}

func (ffs *FastHttpFileServer) StripSlashes() int {
	return ffs.stripSlashes
}

func NewFastHttpFileServer(root string, stripSlashes int) *FastHttpFileServer {
	return &FastHttpFileServer{
		root:         root,
		stripSlashes: stripSlashes,
	}
}

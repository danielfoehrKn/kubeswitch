package gotree 

const (
	newLine      = "\n"
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
)

type (
	tree struct {
		text  string
		items []Tree
	}

	Tree interface {
		Add(text string) Tree
		AddTree(tree Tree)
		Items() []Tree
		Text() string
		Print() string
	}

	printer struct {
	}

	Printer interface {
		Print(Tree) string
	}
)

func New(text string) Tree {
	return &tree{
		text:  text,
		items: []Tree{},
	}
}

func (t *tree) Add(text string) Tree {
	n := New(text)
	t.items = append(t.items, n)
	return n
}

func (t *tree) AddTree(tree Tree) {
	t.items = append(t.items, tree)
}

func (t *tree) Text() string {
	return t.text
}

func (t *tree) Items() []Tree {
	return t.items
}

func (t *tree) Print() string {
	return newPrinter().Print(t)
}

func newPrinter() Printer {
	return &printer{}
}

func (p *printer) Print(t Tree) string {
	return t.Text() + newLine + p.printItems(t.Items(), []bool{})
}

func (p *printer) printText(text string, spaces []bool) string {
	var result string
	last := true
	for _, space := range spaces {
		if space {
			result += emptySpace
		} else {
			result += continueItem
		}
		last = space
	}

	indicator := middleItem
	if last {
		indicator = lastItem
	}

	return result + indicator + text + newLine
}

func (p *printer) printItems(t []Tree, spaces []bool) string {
	var result string
	for i, f := range t {
		last := i == len(t)-1
		result += p.printText(f.Text(), spaces)
		if len(f.Items()) > 0 {
			spacesChild := append(spaces, last)
			result += p.printItems(f.Items(), spacesChild)
		}
	}
	return result
}

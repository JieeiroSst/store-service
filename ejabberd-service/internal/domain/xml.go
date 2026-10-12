package domain

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

const nsXML = "http://www.w3.org/XML/1998/namespace"

type Attr struct {
	Name  string
	Value string
}

type Node interface {
	isNode()
}

type CData string

func (CData) isNode() {}

type Element struct {
	Name     string
	Attrs    []Attr
	Children []Node
}

func (*Element) isNode() {}

func NewElement(name, xmlns string) *Element {
	el := &Element{Name: name}
	if xmlns != "" {
		el.Attrs = []Attr{{Name: "xmlns", Value: xmlns}}
	}
	return el
}

func (e *Element) Attr(name string) string {
	for _, a := range e.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func (e *Element) HasAttr(name string) bool {
	for _, a := range e.Attrs {
		if a.Name == name {
			return true
		}
	}
	return false
}

func (e *Element) SetAttr(name, value string) *Element {
	for i, a := range e.Attrs {
		if a.Name == name {
			e.Attrs[i].Value = value
			return e
		}
	}
	e.Attrs = append(e.Attrs, Attr{Name: name, Value: value})
	return e
}

func (e *Element) RemoveAttr(name string) *Element {
	out := e.Attrs[:0]
	for _, a := range e.Attrs {
		if a.Name != name {
			out = append(out, a)
		}
	}
	e.Attrs = out
	return e
}

func (e *Element) Namespace() string {
	return e.Attr("xmlns")
}

func (e *Element) Append(nodes ...Node) *Element {
	e.Children = append(e.Children, nodes...)
	return e
}

func (e *Element) AppendText(s string) *Element {
	if s != "" {
		e.Children = append(e.Children, CData(s))
	}
	return e
}

func (e *Element) Elements() []*Element {
	var out []*Element
	for _, c := range e.Children {
		if el, ok := c.(*Element); ok {
			out = append(out, el)
		}
	}
	return out
}

func (e *Element) Child(name, xmlns string) *Element {
	for _, c := range e.Children {
		if el, ok := c.(*Element); ok && el.Name == name && (xmlns == "" || el.Namespace() == xmlns) {
			return el
		}
	}
	return nil
}

func (e *Element) ChildrenNamed(name, xmlns string) []*Element {
	var out []*Element
	for _, c := range e.Children {
		if el, ok := c.(*Element); ok && el.Name == name && (xmlns == "" || el.Namespace() == xmlns) {
			out = append(out, el)
		}
	}
	return out
}

func (e *Element) ChildNS(xmlns string) *Element {
	for _, c := range e.Children {
		if el, ok := c.(*Element); ok && el.Namespace() == xmlns {
			return el
		}
	}
	return nil
}

func (e *Element) RemoveChildren(name, xmlns string) *Element {
	out := e.Children[:0]
	for _, c := range e.Children {
		if el, ok := c.(*Element); ok && el.Name == name && (xmlns == "" || el.Namespace() == xmlns) {
			continue
		}
		out = append(out, c)
	}
	e.Children = out
	return e
}

func (e *Element) Text() string {
	var b strings.Builder
	for _, c := range e.Children {
		if t, ok := c.(CData); ok {
			b.WriteString(string(t))
		}
	}
	return b.String()
}

func (e *Element) ChildText(name, xmlns string) string {
	if c := e.Child(name, xmlns); c != nil {
		return c.Text()
	}
	return ""
}

func (e *Element) Copy() *Element {
	if e == nil {
		return nil
	}
	out := &Element{Name: e.Name, Attrs: append([]Attr(nil), e.Attrs...)}
	if len(e.Children) > 0 {
		out.Children = make([]Node, len(e.Children))
		for i, c := range e.Children {
			if el, ok := c.(*Element); ok {
				out.Children[i] = el.Copy()
			} else {
				out.Children[i] = c
			}
		}
	}
	return out
}

func (e *Element) String() string {
	var b strings.Builder
	e.write(&b, "")
	return b.String()
}

func (e *Element) StringNS(parentNS string) string {
	var b strings.Builder
	e.write(&b, parentNS)
	return b.String()
}

func (e *Element) write(b *strings.Builder, parentNS string) {
	b.WriteByte('<')
	b.WriteString(e.Name)
	ns := parentNS
	for _, a := range e.Attrs {
		if a.Name == "xmlns" {
			if a.Value == parentNS {
				continue
			}
			ns = a.Value
		}
		b.WriteByte(' ')
		b.WriteString(a.Name)
		b.WriteString(`="`)
		b.WriteString(EscapeAttr(a.Value))
		b.WriteByte('"')
	}
	if len(e.Children) == 0 {
		b.WriteString("/>")
		return
	}
	b.WriteByte('>')
	for _, c := range e.Children {
		switch n := c.(type) {
		case CData:
			b.WriteString(EscapeText(string(n)))
		case *Element:
			n.write(b, ns)
		}
	}
	b.WriteString("</")
	b.WriteString(e.Name)
	b.WriteByte('>')
}

var (
	attrEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	textEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

func EscapeAttr(s string) string { return attrEscaper.Replace(s) }

func EscapeText(s string) string { return textEscaper.Replace(s) }

func StartElementToElement(se xml.StartElement, parentNS string) *Element {
	el := &Element{Name: se.Name.Local}
	switch {
	case se.Name.Space != "":
		el.Attrs = append(el.Attrs, Attr{Name: "xmlns", Value: se.Name.Space})
	case parentNS != "":
		el.Attrs = append(el.Attrs, Attr{Name: "xmlns", Value: parentNS})
	}
	for _, a := range se.Attr {
		switch {
		case a.Name.Space == "" && a.Name.Local == "xmlns":
		case a.Name.Space == "xmlns":
		case a.Name.Space == nsXML || a.Name.Space == "xml":
			el.Attrs = append(el.Attrs, Attr{Name: "xml:" + a.Name.Local, Value: a.Value})
		default:
			el.Attrs = append(el.Attrs, Attr{Name: a.Name.Local, Value: a.Value})
		}
	}
	return el
}

func ReadElement(d *xml.Decoder, start xml.StartElement, parentNS string) (*Element, error) {
	root := StartElementToElement(start, parentNS)
	stack := []*Element{root}
	for {
		tok, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
		cur := stack[len(stack)-1]
		switch t := tok.(type) {
		case xml.StartElement:
			child := StartElementToElement(t, cur.Namespace())
			cur.Children = append(cur.Children, child)
			stack = append(stack, child)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return root, nil
			}
		case xml.CharData:
			if len(t) > 0 {
				if n := len(cur.Children); n > 0 {
					if prev, ok := cur.Children[n-1].(CData); ok {
						cur.Children[n-1] = prev + CData(t)
						continue
					}
				}
				cur.Children = append(cur.Children, CData(string(t)))
			}
		}
	}
}

func ParseElement(s string) (*Element, error) {
	d := xml.NewDecoder(strings.NewReader(s))
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		if se, ok := tok.(xml.StartElement); ok {
			return ReadElement(d, se, "")
		}
	}
}

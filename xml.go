package properties

import (
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"strings"
)

// xml file header title
const Header = `<!DOCTYPE properties SYSTEM "http://github.com/zooyer/properties">` + "\n"

// interface
type XmlSupport interface {
	load(props *Properties, in io.Reader) error
	store(props *Properties, out io.Writer, comment []byte, encoding string) error
}

// global single SmlSupport object
var PROVIDER = newXmlSupport()

func load(props *Properties, in io.Reader) error {
	return PROVIDER.load(props, in)
}

func save(props *Properties, out io.Writer, comment []byte, encoding string) error {
	return PROVIDER.store(props, out, comment, encoding)
}

// XmlSupport implement
type xmlSupport struct{}

func newXmlSupport() XmlSupport {
	return new(xmlSupport)
}

func (x *xmlSupport) load(props *Properties, in io.Reader) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	var m XMLProperties
	if err = xml.Unmarshal(data, &m); err != nil {
		return err
	}
	p := props.New()
	if err = x.toProperties(p, &m); err != nil {
		return err
	}
	props.Hashtable = p.Hashtable

	return nil
}

func (x *xmlSupport) store(props *Properties, out io.Writer, comment []byte, encoding string) error {
	if props == nil {
		return errors.New("props(Properties) is <nil>")
	}
	if out == nil {
		return errors.New("out(io.Writer) is <nil>")
	}
	switch strings.ToUpper(encoding) {
	case "UTF-8":
	default:
		return errors.New("not support encoding <" + encoding + ">")
	}
	xp, err := x.toXML(props)
	if err != nil {
		return err
	}
	data, err := xp.ToXML(comment)
	if err != nil {
		return err
	}
	if _, err = out.Write([]byte(Header)); err != nil {
		return err
	}
	if _, err = out.Write(data); err != nil {
		return err
	}

	return nil
}

func (x *xmlSupport) toProperties(props *Properties, xp *XMLProperties) error {
	if props == nil {
		return errors.New("props(Properties) is <nil>")
	}
	if xp == nil {
		return errors.New("xp(XMLProperties) is <nil>")
	}

	for _, v := range xp.Entry {
		props.Put(v.Key, v.CDATA)
	}

	return nil
}

func (x *xmlSupport) toXML(props *Properties) (*XMLProperties, error) {
	if props == nil {
		return nil, errors.New("props(Properties) is <nil>")
	}
	var xp = new(XMLProperties)
	xp.Entry = make([]XMLReaderEntry, props.Size())
	for i, key := range props.StringPropertyNames() {
		val, exist := props.GetProperty(key)
		if exist {
			xp.Entry[i] = XMLReaderEntry{
				Key:   key,
				CDATA: val,
			}
		}
	}

	return xp, nil
}

// entity conversion
type XMLReaderEntry struct {
	XMLName xml.Name `xml:"entry"`
	Key     string   `xml:"key,attr"`
	CDATA   string   `xml:",cdata"`
}

type XMLWriterEntry struct {
	XMLName xml.Name `xml:"entry"`
	Key     string   `xml:"key,attr"`
	Data    string   `xml:",chardata"`
}

type XMLReader struct {
	XMLName xml.Name         `xml:"properties"`
	Comment string           `xml:"comment"`
	Entry   []XMLReaderEntry `xml:"entry"`
}

type XMLWriter struct {
	XMLName xml.Name         `xml:"properties"`
	Entry   []XMLWriterEntry `xml:"entry"`
}

type XMLWriterComment struct {
	Comment string `xml:"comment"`
	XMLWriter
}

type XMLProperties struct {
	XMLReader
}

func (x *XMLProperties) ToXML(comments []byte) ([]byte, error) {
	w := new(XMLWriter)
	w.XMLName = x.XMLName
	w.Entry = make([]XMLWriterEntry, len(x.Entry))
	for i, e := range x.Entry {
		w.Entry[i] = XMLWriterEntry{
			XMLName: e.XMLName,
			Key:     e.Key,
			Data:    e.CDATA,
		}
	}

	if comments == nil {
		return xml.MarshalIndent(w, "", "    ")
	}

	wc := new(XMLWriterComment)
	wc.Comment = string(comments)
	wc.XMLWriter = *w

	return xml.MarshalIndent(wc, "", "    ")
}

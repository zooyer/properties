package properties

import (
	"encoding/xml"
	"io/ioutil"
	"testing"
)

func TestXML(t *testing.T) {
	data,err := ioutil.ReadFile("test/test.xml")
	if err != nil {
		t.Fatal(err)
	}
	var m XMLProperties
	if err = xml.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", m)
	x,err := m.ToXML([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(x))
	return
}


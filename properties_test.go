package properties

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"strings"
	"testing"
)

func TestNewLineReader(t *testing.T) {
	const lines = `line1
line2
# comment
!zzzzzz
line3
         
	 line4
line5
`
	tests := []string{"line1", "line2", "line3", "line4", "line5"}
	lr := NewLineReader(strings.NewReader(lines))
	for i := range tests {
		n := lr.readLine()
		diff := cmp.Diff(lr.lineBuf[:n], []byte(tests[i]))
		if diff != "" {
			t.Fatal(i, diff)
		}
	}
	diff := cmp.Diff(lr.readLine(), -1)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestWriteComments(t *testing.T) {
	var err error
	var buf strings.Builder
	const comments = "comment1\n!comment2\n#comment3"
	const result = "#comment1\r\n!comment2\r\n#comment3\r\n"
	if err = writeComments(&buf, []byte(comments)); err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(buf.String(), result)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestToHix(t *testing.T) {
	tests := map[int]byte{
		0:  '0',
		1:  '1',
		2:  '2',
		3:  '3',
		4:  '4',
		5:  '5',
		6:  '6',
		7:  '7',
		8:  '8',
		9:  '9',
		10: 'A',
		11: 'B',
		12: 'C',
		13: 'D',
		14: 'E',
		15: 'F',
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			diff := cmp.Diff(toHex(i), test)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestNewProperties(t *testing.T) {
	var p = NewProperties()
	p.SetProperty("title", "test")
	p.SetProperty("version", "v1.1.1")
	p.SetProperty("test", "")
	p.Remove("test")
	p.List(os.Stdout)
}

func TestProperties_Load(t *testing.T) {
	var p = NewProperties()

	file, err := os.Open("test/log4j2.properties")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = p.Load(file); err != nil {
		t.Fatal(err)
	}

	p.List(os.Stdout)
}

func TestProperties_LoadFromXML(t *testing.T) {
	var p = NewProperties()
	file, err := os.Open("test/log4j2.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = p.LoadFromXML(file); err != nil {
		t.Fatal(err)
	}
	fmt.Println("size:", p.Size())
	fmt.Println(p.GetProperty("status"))
	p.Remove("status")
	fmt.Println(p.GetProperty("status"))
	fmt.Println(p.StringPropertyNames())

	p.Store(os.Stdout, []byte("-----------comment-----------"))
}

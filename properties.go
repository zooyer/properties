package properties

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"runtime"
	"sync"
	"time"
)

// A table of hex digits
var hexDigit = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

type Properties struct {
	Hashtable
	mutex sync.Mutex

	// A property list that contains default values for any keys not
	// found in this property list.
	defaults *Properties
}

// Creates an empty property list with no default values.
func NewProperties() *Properties {
	var hash = NewHashtable()
	return &Properties{
		Hashtable: hash,
		defaults:  nil,
	}
}

// Creates an empty property list with the specified defaults.
func NewPropertiesDefault(defaults *Properties) *Properties {
	var properties = NewProperties()
	properties.defaults = defaults

	return properties
}

// Calls the Hashtable method Put. Provided for parallelism with the
// Get method. Enforces use of strings for property keys and values.
// The value returned is the result of the Hashtable call to Put.
func (p *Properties) SetProperty(key, value string) interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.Put(key, value)
}

// The specified Reader remains open after this method returns.
// reader the input character reader.
func (p *Properties) Load(reader io.Reader) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.load0(NewLineReader(reader))
}

func (p *Properties) load0(lr *LineReader) error {
	var convertBuf = make([]byte, 4096)
	var limit, keyLen, valueStart int
	var c byte
	var hasSep, precedingBackslash bool

	for limit = lr.readLine(); limit >= 0; limit = lr.readLine() {
		c = 0
		keyLen = 0
		valueStart = limit
		hasSep = false

		//fmt.Println("line=<" + string(lr.lineBuf[:limit]) + ">")
		precedingBackslash = false
		for keyLen < limit {
			c = lr.lineBuf[keyLen]
			//need check if escaped.
			if (c == '=' || c == ':') && !precedingBackslash {
				valueStart = keyLen + 1
				hasSep = true
				break
			} else if (c == ' ' || c == '\t' || c == '\f') && !precedingBackslash {
				valueStart = keyLen + 1
				break
			}
			if c == '\\' {
				precedingBackslash = !precedingBackslash
			} else {
				precedingBackslash = false
			}
			keyLen++
		}
		for valueStart < limit {
			c = lr.lineBuf[valueStart]
			if c != ' ' && c != '\t' && c != '\f' {
				if !hasSep && (c == '=' || c == ':') {
					hasSep = true
				} else {
					break
				}
			}
			valueStart++
		}

		key, err := p.loadConvert(lr.lineBuf, 0, keyLen, convertBuf)
		if err != nil {
			return err
		}
		value, err := p.loadConvert(lr.lineBuf, valueStart, limit-valueStart, convertBuf)
		if err != nil {
			return err
		}
		p.Put(key, value)
	}

	return nil
}

// Converts encoded &#92;uxxxx to unicode chars
// and changes special saved chars to their original forms
func (p *Properties) loadConvert(in []byte, off, length int, convertBuf []byte) (string, error) {
	if len(convertBuf) < length {
		var newLen = length * 2
		if newLen < 0 {
			newLen = int(^uint(0) >> 1)
		}
		convertBuf = make([]byte, newLen)
	}

	var aChar byte
	var out = convertBuf
	var outLen = 0
	var end = off + length

	for off < end {
		aChar = in[off]
		off++
		if aChar == '\\' {
			aChar = in[off]
			off++
			if aChar == 'u' {
				// Read the xxxx
				var value = 0

				for i := 0; i < 4; i++ {
					aChar = in[off]
					off++
					switch aChar {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						value = (value << 4) + int(aChar) - '0'
					case 'a', 'b', 'c', 'd', 'e', 'f':
						value = (value << 4) + 10 + int(aChar) - 'a'
					case 'A', 'B', 'C', 'D', 'E', 'F':
						value = (value << 4) + 10 + int(aChar) - 'A'
					default:
						return "", errors.New("malformed \\uxxxx encoding")
					}
				}
				out[outLen] = byte(value)
				outLen++
			} else {
				if aChar == 't' {
					aChar = '\t'
				} else if aChar == 'r' {
					aChar = '\r'
				} else if aChar == 'n' {
					aChar = '\n'
				} else if aChar == 'f' {
					aChar = '\f'
				}
				out[outLen] = aChar
				outLen++
			}
		} else {
			out[outLen] = aChar
			outLen++
		}
	}

	return string(out[:outLen]), nil
}

// Converts unicode to encoded &#92;uxxxx and escapes
// special characters with a preceding slash
func (p *Properties) saveConvert(theString string, escapeSpace, escapeUnicode bool) string {
	var length = len(theString)
	var bufLen = length * 2
	if bufLen < 0 {
		bufLen = int(^uint(0) >> 1)
	}

	var outBuffer bytes.Buffer

	for x := 0; x < length; x++ {
		var aChar = theString[x]
		// Handle common case first, selecting largest block that
		// avoids the specials below
		if (aChar > 61) && (aChar < 127) {
			if aChar == '\\' {
				outBuffer.WriteByte('\\')
				outBuffer.WriteByte('\\')
				continue
			}
			outBuffer.WriteByte(aChar)
			continue
		}
		switch aChar {
		case ' ':
			if x == 0 || escapeSpace {
				outBuffer.WriteByte('\\')
			}
			outBuffer.WriteByte(' ')
		case '\t':
			outBuffer.WriteByte('\\')
			outBuffer.WriteByte('t')
		case '\n':
			outBuffer.WriteByte('\\')
			outBuffer.WriteByte('n')
		case '\r':
			outBuffer.WriteByte('\\')
			outBuffer.WriteByte('r')
		case '\f':
			outBuffer.WriteByte('\\')
			outBuffer.WriteByte('f')
		case '=':
			fallthrough
		case ':', '#', '!':
			outBuffer.WriteByte('\\')
			outBuffer.WriteByte(aChar)
		default:
			if ((aChar < 0x0020) || (aChar > 0x007e)) && escapeUnicode {
				outBuffer.WriteByte('\\')
				outBuffer.WriteByte('u')
				outBuffer.WriteByte(toHex(int(aChar>>12) & 0xF))
				outBuffer.WriteByte(toHex(int(aChar>>8) & 0xF))
				outBuffer.WriteByte(toHex(int(aChar>>4) & 0xF))
				outBuffer.WriteByte(toHex(int(aChar) & 0xF))
			} else {
				outBuffer.WriteByte(aChar)
			}
		}
	}

	return outBuffer.String()
}

// This method does not return error or panic
// if an I/O error occurs while saving the property list.
// Deprecated
func (p *Properties) Save(writer io.Writer, comments []byte) {
	defer func() { recover() }()
	_ = p.Store(writer, comments)
}

// Writes this property list (key and element pairs) in this Properties table to the
// output character stream in a format suitable for using the io.Reader load(Reader)
// After the entries have been written, the output stream is flushed.
// The output stream remains open after this method returns.
func (p *Properties) Store(writer io.Writer, comments []byte) error {
	return p.store0(bufio.NewWriter(writer), comments, true)
}

func (p *Properties) store0(w io.Writer, comments []byte, escUnicode bool) (err error) {
	var bw = bufio.NewWriter(w)
	if comments != nil {
		if err = writeComments(bw, comments); err != nil {
			return err
		}
	}
	if _, err = bw.WriteString("# " + time.Now().Format(time.UnixDate)); err != nil {
		return err
	}
	if _, err = bw.Write(newLine()); err != nil {
		return err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, key := range p.Keys() {
		var val = p.Get(key)

		var sKey = key.(string)
		var sVal = val.(string)

		sKey = p.saveConvert(sKey, true, escUnicode)
		// No need to escape embedded and trailing spaces for value, hence
		// pass false to flag.
		sVal = p.saveConvert(sVal, false, escUnicode)
		if _, err = bw.WriteString(sKey + " = " + sVal); err != nil {
			return err
		}
		if _, err = bw.Write(newLine()); err != nil {
			return err
		}
	}

	return bw.Flush()
}

// Loads all of the properties represented by the XML document on the
// specified input stream into this properties table.
// An implementation is required to read XML documents that use the
// UTF-8 or UTF-16 encoding.
// An implementation may support additional encodings.
func (p *Properties) LoadFromXML(reader io.Reader) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return load(p, reader)
}

// Call p.StoreToXMLByEncoding(writer, comment, "UTF-8").
func (p *Properties) StoreToXML(writer io.Writer, comments []byte) error {
	return p.StoreToXMLByEncoding(writer, comments, "UTF-8")
}

// The specified writer remains open after this method returns.
// Emits an XML document representing all of the properties contained in this table.
// An invocation of this method of the form p.StoreToXML(writer, comment)
// behaves in exactly the same way as the invocation.
func (p *Properties) StoreToXMLByEncoding(writer io.Writer, comments []byte, encoding string) error {
	return save(p, writer, comments, encoding)
}

// Searches for the property with the specified key in this property list.
// If the key is not found in this property list, the default property list,
// and its defaults, recursively, are then checked. The method returns
// Return "", false if the property is not found.
func (p *Properties) GetProperty(key string) (val string, exist bool) {
	var oval = p.Get(key)
	if sVal, ok := oval.(string); ok {
		return sVal, true
	}

	if p.defaults != nil {
		return p.defaults.GetProperty(key)
	}

	return "", false
}

// Searches for the property with the specified key in this property list.
// If the key is not found in this property list, the default property list,
// and its defaults, recursively, are then checked. The method returns the
// default value argument if the property is not found.
func (p *Properties) GetPropertyByDefault(key, defaultValue string) string {
	value, exist := p.GetProperty(key)
	if !exist {
		return defaultValue
	}

	return value
}

// Returns an enumeration of all the keys in this property list,
// including distinct keys in the default property list if a key
// of the same name has not already been found from the main
// properties list.
// Return  an enumeration of all the keys in this property list, including
// the keys in the default property list.
func (p *Properties) PropertyNames() []interface{} {
	var h = p.newHashtable()
	p.enumerate(h)

	return h.Keys()
}

// Returns a set of keys in this property list where
// the key and its corresponding value are strings,
// including distinct keys in the default property list if a key
// of the same name has not already been found from the main
// properties list. Properties whose key or value is not
// of type string are omitted.
// The returned set is not backed by the Properties object.
// Changes to this Properties are not reflected in the set,
// or vice versa.
func (p *Properties) StringPropertyNames() []string {
	var h = p.newHashtable()
	p.enumerateStringProperties(h)

	var set = make([]string, 0, h.Size())
	for _, key := range h.Keys() {
		set = append(set, key.(string))
	}

	return set
}

// Rather than use an anonymous inner class to share common code, this
// method is duplicated in order to ensure that a non-1.1 compiler can
// compile this file.
func (p *Properties) List(out io.Writer) {
	_, _ = out.Write(append([]byte("-- listing properties --"), newLine()...))

	var h = p.newHashtable()
	p.enumerate(h)
	for _, key := range h.Keys() {
		var val = h.Get(key)
		var sKey = key.(string)
		var sVal = val.(string)

		if len(sVal) > 40 {
			sVal = string([]byte(sVal)[:37]) + "..."
		}
		_, _ = out.Write(append([]byte(sKey+" = "+sVal), newLine()...))
	}
}

// Enumerates all key/value pairs in the specified hashtable.
func (p *Properties) enumerate(h Hashtable) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.defaults != nil {
		p.defaults.enumerate(h)
	}

	for _, key := range p.Hashtable.Keys() {
		h.Put(key, p.Hashtable.Get(key))
	}
}

// Enumerates all key/value pairs in the specified hashtable
// and omits the property if the key or value is not a string.
func (p *Properties) enumerateStringProperties(h Hashtable) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.defaults != nil {
		p.defaults.enumerateStringProperties(h)
	}

	// safe assert type string
	for _, key := range p.Hashtable.Keys() {
		var val = p.Hashtable.Get(key)
		if _, ok := key.(string); ok {
			if _, ok = val.(string); ok {
				h.Put(key.(string), val.(string))
			}
		}
	}
}

// Create a hashtable that is the same type as itself.
func (p *Properties) newHashtable() Hashtable {
	return p.Hashtable.New()
}

// Properties to map.
func (p *Properties) ToMap() map[interface{}]interface{} {
	var m = make(map[interface{}]interface{})
	keys := p.Keys()
	for _, key := range keys {
		m[key] = p.Get(key)
	}

	return m
}

// Create a Properties that is the same type as itself.
func (p *Properties) New() *Properties {
	prop := NewProperties()
	prop.Hashtable = p.newHashtable()

	return prop
}

// Convert a nibble to a hex character
// param nibble the nibble to convert.
func toHex(nibble int) byte {
	return hexDigit[nibble&0xF]
}

// Write a comments.
func writeComments(w io.Writer, comments []byte) (err error) {
	if comments == nil {
		return nil
	}
	bw := bufio.NewWriter(w)
	defer func() {
		if err == nil {
			err = bw.Flush()
		}
	}()
	if err = bw.WriteByte('#'); err != nil {
		return err
	}
	var length = len(comments)
	var current = 0
	var last = 0
	var uu = make([]byte, 6)
	uu[0] = '\\'
	uu[1] = 'u'
	for current < length {
		var c = comments[current]
		if c > '\u00ff' || c == '\n' || c == '\r' {
			if last != current {
				if _, err = bw.Write([]byte(comments)[last:current]); err != nil {
					return err
				}
			}
			if c > '\u00ff' {
				uu[2] = toHex(int(c>>12) & 0xf)
				uu[3] = toHex(int(c>>8) & 0xf)
				uu[4] = toHex(int(c>>4) & 0xf)
				uu[5] = toHex(int(c) & 0xf)
				if _, err = bw.Write(uu); err != nil {
					return err
				}
			} else {
				if _, err = bw.Write(newLine()); err != nil {
					return err
				}
				if c == '\r' && current != length-1 && comments[current+1] == '\n' {
					current++
				}
				if current == length-1 || comments[current+1] != '#' && comments[current+1] != '!' {
					if err = bw.WriteByte('#'); err != nil {
						return err
					}
				}
			}
			last = current + 1
		}
		current++
	}

	if last != current {
		if _, err = bw.Write([]byte(comments)[last:current]); err != nil {
			return err
		}
	}
	if _, err = bw.Write(newLine()); err != nil {
		return err
	}

	return nil
}

// Return a line separator.  The line separator string is defined by the
// system property line.separator, and is not necessarily a single
// newline ('\n') or ("\r\n") character slice.
func newLine() []byte {
	const (
		CR = "\r"
		LF = "\n"
	)
	switch runtime.GOOS {
	case "windows":
		return []byte(CR + LF)
	case "linux":
		fallthrough
	default:
		return []byte(LF)
	}
}

// Read in a "logical line" from an Reader, skip all comment and blank
// lines and filter out those leading whitespace characters
// (\u0020, \u0009 and \u000c) from the beginning of a "natural line".
// Method returns the char length of the "logical line" and stores
// the line in "lineBuf".

type LineReader struct {
	inByteBuf []byte
	lineBuf   []byte

	inLimit int
	inOff   int

	reader io.Reader
}

func NewLineReader(reader io.Reader) *LineReader {
	return &LineReader{
		inByteBuf: make([]byte, 8192),
		lineBuf:   make([]byte, 1024),
		reader:    reader,
		inLimit:   0,
		inOff:     0,
	}
}

func (l *LineReader) readLine() int {
	var length = 0
	var c byte = 0
	var (
		skipWhiteSpace     = true
		isCommentLine      = false
		isNewLine          = true
		appendedLineBegin  = false
		precedingBackslash = false
		skipLF             = false
	)

	for true {
		if l.inOff >= l.inLimit {
			n, err := l.reader.Read(l.inByteBuf)
			l.inLimit = n
			l.inOff = 0
			if err != nil || l.inLimit <= 0 {
				if length == 0 || isCommentLine {
					return -1
				}
				if precedingBackslash {
					length--
				}
				return length
			}
		}

		c = l.inByteBuf[l.inOff]
		l.inOff++

		if skipLF {
			skipLF = false
			if c == '\n' {
				continue
			}
		}
		if skipWhiteSpace {
			if c == ' ' || c == '\t' || c == '\f' {
				continue
			}
			if !appendedLineBegin && (c == '\r' || c == '\n') {
				continue
			}
			skipWhiteSpace = false
			appendedLineBegin = false
		}
		if isNewLine {
			isNewLine = false
			if c == '#' || c == '!' {
				isCommentLine = true
				continue
			}
		}

		if c != '\n' && c != '\r' {
			l.lineBuf[length] = c
			length++
			if length == len(l.lineBuf) {
				var newLength = length * 2
				if newLength < 0 {
					newLength = int(^uint(0) >> 1)
				}
				var buf = make([]byte, newLength)
				copy(buf, l.lineBuf)
				l.lineBuf = buf
			}
			//flip the preceding backslash flag
			if c == '\\' {
				precedingBackslash = !precedingBackslash
			} else {
				precedingBackslash = false
			}
		} else {
			// reached EOL
			if isCommentLine || length == 0 {
				isCommentLine = false
				isNewLine = true
				skipWhiteSpace = true
				length = 0
				continue
			}
			if l.inOff >= l.inLimit {
				n, err := l.reader.Read(l.inByteBuf)
				l.inLimit = n
				l.inOff = 0
				if err != nil || l.inLimit <= 0 {
					if precedingBackslash {
						length--
					}
					return length
				}
				if precedingBackslash {
					length -= 1
					//skip the leading whitespace characters in following line
					skipWhiteSpace = true
					appendedLineBegin = true
					precedingBackslash = false
					if c == '\r' {
						skipLF = true
					}
				} else {
					return length
				}
			}
			if precedingBackslash {
				length -= 1
				//skip the leading whitespace characters in following line
				skipWhiteSpace = true
				appendedLineBegin = true
				precedingBackslash = false
				if c == '\r' {
					skipLF = true
				}
			} else {
				return length
			}
		}
	}

	return -1
}

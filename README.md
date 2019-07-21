# The Go Properties

![](https://travis-ci.org/boennemann/badges.svg?branch=master)  ![](https://img.shields.io/badge/license-MIT-blue.svg)  ![](https://img.shields.io/badge/godoc-reference-blue.svg)

Go Properties is migrate java Properties.

![](https://github.com/golang/go/blob/master/doc/gopher/fiveyears.jpg?raw=true)


#### Download and Install

```shell
go get github.com/zooyer/properties
```

#### Features

- 100% compatible with java Properties
- xml properties support
- persistence support

#### Example

1. set/get/list

   ```go
   package main
   
   import (
   	"fmt"
   	"github.com/zooyer/properties"
   )
   
   func main() {
   	prop := properties.NewProperties()
   
   	prop.SetProperty("title", "properties")
   	prop.SetProperty("language", "golang")
   	prop.SetProperty("version", "1.9.2")
   
   	fmt.Println(prop.GetProperty("title"))
   	fmt.Println(prop.GetProperty("language"))
   	fmt.Println(prop.GetProperty("version"))
   
   	prop.List(os.Stdout)
   }
   ```

   output:

   ```shell
   properties true
   golang true
   1.9.2 true
   -- listing properties --
   title = properties
   language = golang
   version = 1.9.2
   ```

2. store

   ```go
   package main
   
   import (
   	"fmt"
   	"github.com/zooyer/properties"
   	"os"
   )
   
   func main() {
   	prop := properties.NewProperties()
   
   	prop.SetProperty("title", "properties")
   	prop.SetProperty("language", "golang")
   	prop.SetProperty("version", "1.9.2")
   
   	file, err := os.Create("test.properties")
   	if err != nil {
   		panic(err)
   	}
   	defer file.Close()
   
   	if err = prop.Store(file, []byte("golang properties test comment")); err != nil {
   		panic(err)
   	}
   	file.Seek(0, io.SeekStart)
   
   	prop2 := properties.NewProperties()
   	if err = prop2.Load(file); err != nil {
   		panic(err)
   	}
   	prop2.List(os.Stdout)
   }
   ```

   output:

   ```shell
   -- listing properties --
   title = properties
   language = golang
   version = 1.9.2
   ```

   test.propertes:

   ```shell
   #golang properties test comment
   # Sun Jul 21 12:42:11 CST 2019
   language = golang
   version = 1.9.2
   title = properties
   ```

3. xml

   ```go
   package main
   
   import (
   	"github.com/zooyer/properties"
   	"os"
   )
   
   func main() {
   	prop := properties.NewProperties()
   
   	prop.SetProperty("title", "properties")
   	prop.SetProperty("language", "golang")
   	prop.SetProperty("version", "1.9.2")
   
   	prop.StoreToXMLByEncoding(os.Stdout, []byte("test xml comment"), "utf-8")
   }
   ```

   output:

   ```shell
   <!DOCTYPE properties SYSTEM "http://github.com/zooyer/properties">
   <properties>
       <comment>test xml comment</comment>
       <entry key="version">1.9.2</entry>
       <entry key="title">properties</entry>
       <entry key="language">golang</entry>
   </properties>
   ```


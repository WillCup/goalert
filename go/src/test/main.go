package main
import (
    "fmt"
    // . "github.com/apaxa-go/eval"
    // constanth "github.com/apaxa-go/helper/goh/constanth"
    // "math"
    "encoding/json"
    "net/http"
    "io/ioutil"
)

type exampleString string

func (s exampleString) String() exampleString { return "!" + s + "!" }

type exampleStruct struct {
    A, B int
}

func (s exampleStruct) Sum() int { return s.A + s.B }


// 解析http返回的数据
func get_result(byt []byte, exp string) bool {
    var dat map[string]interface{}
    if err := json.Unmarshal(byt, &dat); err != nil {
        panic(err)
    }
    // go语言不能像python的json.dump一样直接将json转为map嵌套结构，要一层一层弄。
    jsonObj, ok := dat["statusInfo"].(map[string]interface{})
    if !ok {
        fmt.Println("Not Ok.")
    } else {
        fmt.Println(jsonObj["global"])
        fmt.Println("成功")
    }  

    fmt.Println(dat["statusInfo"])
    return false
}

type Ressult struct{
    status int
    statusInfo map[string]interface{}
}

func httpGet() {
    resp, err := http.Get("http://will_api.will.com.cn/will_api_test.cgi?signKey=d28105da-4703-11e9-9a0f-faxxxxe0xxxx5c&startTime=20190102&endTime=20190103&series_id=69")
    if err != nil {
        // handle error
        fmt.Errorf("%s", "发生了错误")
    }
 
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        // handle error
        fmt.Errorf("%s", "发生了错误")
    }
 
    fmt.Println(string(body))
    // var ss Ressult
    // json.Unmarshal(body, &ss)
    // fmt.Printf("Results: %v\n", ss.statusInfo)
    get_result(body, "xxx")
}

func main() {
    fmt.Println("hello go")
    httpGet()
    // 简单的例子
    // src:="int8(1*(1+2))"
    // expr,err:=ParseString(src,"")
    // if err != nil {
    //     fmt.Println("发生错误")
    // }
    // r,err:=expr.EvalToInterface(nil)
    // if err != nil {
    //     fmt.Println("发生错误")
    // }
    // fmt.Printf("%v %T", r, r)   // "3 int8"

// 复杂的例子
    // c := make(chan int64, 10)
    // c <- 2

    // src := `exampleString(fmt.Sprint(interface{}(math.MaxInt32/exampleStruct(struct{ A, B int }{3, 5}).Sum()+int(<-(<-chan int64)(c))-cap(make([]string, 1, 100))))).String().String() + "."`

    // expr, err := ParseString(src, "")
    // if err != nil {
    //     return
    // }
    // a := Args{
    //     "exampleString": MakeTypeInterface(exampleString("")),
    //     "fmt.Sprint":    MakeDataRegularInterface(fmt.Sprint),
    //     "math.MaxInt32": MakeDataUntypedConst(constanth.MakeUint(math.MaxInt32)),
    //     "exampleStruct": MakeTypeInterface(exampleStruct{}),
    //     "c":             MakeDataRegularInterface(c),
    // }
    // r, err := expr.EvalToInterface(a)
    // if err != nil {
    //     return
    // }
    // fmt.Printf("%v %T\n", r, r) // "!!268435357!!. eval.exampleString"
    // return
}
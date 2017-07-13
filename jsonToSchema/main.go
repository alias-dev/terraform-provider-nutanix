package main


import (
	"bufio"
	"github.com/ideadevice/terraform-ahv-provider-plugin/jsonToSchema/gopath"
	"bytes"
	"strings"
	"flag"
	"fmt"
	glog "log"
	"os"
)

var (
	configFilePath = gopath.ConfigPath + "/virtualmachineconfig.autogenerated.go"
	schemaFilePath = gopath.SchemaPath + "/virtualmachineschema.autogenerated.go"
	log               = glog.New(os.Stderr, "", glog.Lshortfile)
	fstruct           = flag.String("structName", "VmIntentInput", "struct name for json object")
	debug             = false
	fileSchema, _ 	  = os.Create(os.ExpandEnv(schemaFilePath))
	wSchema			  = bufio.NewWriter(fileSchema)
	depth 			  = 0
)

func tabN(n int ){
	i := 0
	for i<n {
		i++
		fmt.Fprintf(wSchema, "\t")
	}
}

func main() {
	fmt.Fprintf(wSchema, "%s\n", schemaHeader)
	depth = 2
	_, _, err := xreflect("VmIntentInput")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(wSchema, "\t}\n}")
	wSchema.Flush()
	fileSchema.Close()
}

func xreflect(name string) ([]byte, []byte, error) {
	var (
		bufConfig = new(bytes.Buffer)
		bufList = new(bytes.Buffer)
	)
	file ,err := os.Open(os.ExpandEnv( gopath.StructPath + fromCamelcase(structNameMap[name]) +".go"))
	if err != nil {
		log.Fatal(err)
	}
    defer file.Close()
    scanner := bufio.NewScanner(file)
    maps := make(map[string]string)
    var flag bool
    for scanner.Scan() {
        words := strings.Fields(scanner.Text())
        if scanner.Text() == "}" {
            break
        }
        if len(words) > 2 {
            if  words[2] == "struct" && words[3] == "{" {
                flag = true
                continue
            }
        }
        if flag && len(words) >1 {
            if words[0] != "//" {
                maps[words[0]] = words[1]
            }
        }
    }

	for key, val := range maps {
		tabN(depth)
		fmt.Fprintf(wSchema, "\"%s\": &schema.Schema{\n", fromCamelcase(key))
		tabN(depth+1)
		fmt.Fprintf(wSchema, "Optional: true,\n")
		switch val {
		case "int64":
			tabN(depth+1)
			fmt.Fprintf(wSchema, "Type: schema.TypeInt,\n")
			fmt.Fprintf(bufConfig, "\t\t\t%s:\t\tconvertToInt(s[\"%s\"]),\n", key, fromCamelcase(key))
		case "string":
			tabN(depth+1)
			fmt.Fprintf(wSchema, "Type: schema.TypeString,\n")
			fmt.Fprintf(bufConfig, "\t\t\t%s:\t\tconvertToString(s[\"%s\"]),\n", key, fromCamelcase(key))
		case "time.Time":
			tabN(depth+1)
			fmt.Fprintf(wSchema, "Type: schema.TypeString,\n")
			fmt.Fprintf(bufConfig, "\t\t\t%s:\t\t%s,\n", key, goFunc(key))
			fmt.Fprintf(bufList, configTime, goFunc(key), goFunc(key), fromCamelcase(key), goFunc(key), goFunc(key), goFunc(key), goFunc(key))
		case "map[string]string":
			tabN(depth+1)
			fmt.Fprintf(wSchema, "Type: schema.TypeMap,\n")
			tabN(depth+1)
			fmt.Fprintf(wSchema, "Elem:     &schema.Schema{Type: schema.TypeString},\n")
			fmt.Fprintf(bufConfig, "\t\t\t%s:\t\tSet%s(s),\n", toCamelcase(key) ,goFunc(key))
			NewField(key, "map[string]string", nil, nil)
		default:
			tabN(depth+1)
			if strings.HasPrefix(val, "[]") {
				val = strings.TrimPrefix(val, "[]")
				fmt.Fprintf(wSchema, "Type: schema.TypeList,\n")
				fmt.Fprintf(bufConfig, "\t\t\t%s:\t\t%s,\n", key, goFunc(key))
				fmt.Fprintf(bufList, configList, goFunc(key), val , fromCamelcase(key), fromCamelcase(key), goFunc(key), fromCamelcase(key), goFunc(key), goFunc(key))
			} else {
				fmt.Fprintf(wSchema, "Type: schema.TypeSet,\n")
				fmt.Fprintf(bufConfig, "\t\t\t%s:\t\tSet%s(s[\"%s\"].(*schema.Set).List(), 0),\n", key, goFunc(key), fromCamelcase(key))
			}
	
			structNameMap[key] = val
			tabN(depth+1)
			fmt.Fprintf(wSchema, "Elem: &schema.Resource{\n")
			tabN(depth+2)
			fmt.Fprintf(wSchema, "Schema: map[string]*schema.Schema{\n")
			depth = depth + 3

			bConfig, bList, err := xreflect(key)
			if err != nil {
				log.Println(err)
				return nil, nil, err
			}
			NewField(key, "struct", bConfig, bList)
			depth = depth - 3 
			tabN(depth+2)
			fmt.Fprintf(wSchema, "},\n")
			tabN(depth+1)
			fmt.Fprintf(wSchema, "},\n")
		}	

		tabN(depth)
		fmt.Fprintf(wSchema, "},\n")
	}
	return bufConfig.Bytes(), bufList.Bytes(), nil
}

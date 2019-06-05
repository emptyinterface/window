package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/emptyinterface/window"
)

type (
	TemplateSet struct {
		me        sync.RWMutex
		templates *template.Template
	}
)

var (
	templateFiles = []string{
		"templates/index.html",
		"templates/_cloudwatch_errors.html",

		"templates/_region.html",
		"templates/_az_sm.html",
		"templates/_autoscaling_group_sm.html",

		"templates/_classic.html",

		"templates/_vpc.html",
		"templates/_vpc_sm.html",

		"templates/_network.html",
		"templates/_network_cgws.html",
		"templates/_network_enis.html",
		"templates/_network_igws.html",
		"templates/_network_vpces.html",
		"templates/_network_vpcps.html",
		"templates/_network_vpgs.html",
		"templates/_network_vpns.html",

		"templates/_cgw.html",
		"templates/_cgw_sm.html",
		"templates/_eni.html",
		"templates/_eni_sm.html",
		"templates/_igw.html",
		"templates/_igw_sm.html",
		"templates/_vpce.html",
		"templates/_vpce_sm.html",
		"templates/_vpcp.html",
		"templates/_vpcp_sm.html",
		"templates/_vpg.html",
		"templates/_vpg_sm.html",
		"templates/_vpn.html",
		"templates/_vpn_sm.html",
		"templates/_natgw.html",
		"templates/_natgw_sm.html",

		"templates/_ami.html",
		"templates/_ami_sm.html",
		"templates/_amis.html",

		"templates/_cloudwatch_alarm.html",
		"templates/_cloudwatch_alarm_sm.html",
		"templates/_cloudwatch_alarms.html",

		"templates/_ecc.html",
		"templates/_ecc_sm.html",
		"templates/_eccs.html",

		"templates/_elb.html",
		"templates/_elb_sm.html",
		"templates/_elbs.html",

		"templates/_instance.html",
		"templates/_instance_sm.html",
		"templates/_instances.html",

		"templates/_lambda.html",
		"templates/_lambda_sm.html",
		"templates/_lambdas.html",

		"templates/_rds.html",
		"templates/_rds_sm.html",
		"templates/_rdss.html",

		"templates/_security_group.html",
		"templates/_security_group_sm.html",
		"templates/_security_groups.html",

		"templates/_sns.html",
		"templates/_sns_sm.html",
		"templates/_snss.html",

		"templates/_sqs.html",
		"templates/_sqs_sm.html",
		"templates/_sqss.html",

		"templates/_subnet.html",
		"templates/_subnet_sm.html",

		"templates/_ami_data.html",
		"templates/_cloudwatch_alarm_data.html",
		"templates/_ecc_data.html",
		"templates/_elb_data.html",
		"templates/_eni_data.html",
		"templates/_instance_data.html",
		"templates/_lambda_data.html",
		"templates/_rds_data.html",
		"templates/_security_group_data.html",
		"templates/_sns_data.html",
		"templates/_sqs_data.html",
	}

	templateFuncs = template.FuncMap{
		"contains":  strings.Contains,
		"titleCase": strings.Title,
		"softTrue":  soft_true,
		// make if a little easier to use with *aws values
		"value": func(v interface{}) interface{} {

			val := reflect.ValueOf(v)

			// resolve the pointer if it's valid
			for val.IsValid() && val.Kind() == reflect.Ptr && !val.IsNil() {
				val = val.Elem()
			}

			return val.Interface()

		},

		"default": func(v interface{}, vv interface{}) interface{} {
			if soft_true(v) {
				return v
			}
			return vv
		},

		"puthtml": func(s string) template.HTML {
			return template.HTML(s)
		},
		"shortTime": func(t time.Time) string {
			return t.Local().Format("Mon, 02 Jan 2006 03:04:05 PM MST")
		},
		"uptime": func(t time.Time) string {
			const day = time.Hour * 24
			dur := (time.Now().Sub(t) / time.Second) * time.Second
			if dur < day {
				return dur.String()
			}
			d := (dur / time.Minute) * time.Minute
			return strings.TrimSuffix(fmt.Sprintf("%dd%s", d/day, d%day), "0s")
		},
		"inSlice": func(needle string, haystack []string) bool {
			for _, hay := range haystack {
				if needle == hay {
					return true
				}
			}
			return false
		},
		"elbInSubnet": func(elb *window.ELB, subnet *window.Subnet) bool {
			for _, inst := range subnet.Instances {
				if inst.ELB == elb {
					return true
				}
			}
			return false
		},
		"elbInstancesInSubnet": func(elb *window.ELB, subnet *window.Subnet) []*window.Instance {
			var insts []*window.Instance
			for _, inst := range subnet.Instances {
				if inst.ELB == elb {
					insts = append(insts, inst)
				}
			}
			return insts
		},
		"jsonMarshal": func(v interface{}) string {
			data, err := json.Marshal(v)
			if err != nil {
				return err.Error()
			}
			return string(data)
		},
		"softEquals": func(v, vv interface{}) bool {
			// fucked  up
			vb, _ := json.Marshal(v)
			vvb, _ := json.Marshal(vv)
			return bytes.Equal(vb, vvb)
		},
		"percent": func(f float64) int {
			return int(f * 100)
		},
		"countState": func(insts []*window.Instance, state string) int {
			var ct int
			for _, inst := range insts {
				if inst.InstanceState != nil && aws.StringValue(inst.InstanceState.Name) == state {
					ct++
				}
			}
			return ct
		},
		"same": func(a, b interface{}) bool {
			return reflect.ValueOf(a) == reflect.ValueOf(b)
		},
		"humanBytes": func(b int64, precision int) string {
			const divisor = 1024
			var i int
			var n = float64(b)
			for ; n >= divisor; i, n = i+1, n/divisor {
			}
			return strings.TrimSuffix(strconv.FormatFloat(n, 'f', precision, 64), ".0") + sizeNames[i]
		},
		"alarmState": func(alarms []*window.CloudWatchAlarm, state string) []*window.CloudWatchAlarm {
			var as []*window.CloudWatchAlarm
			for _, alarm := range alarms {
				if alarm.StateValue == state {
					as = append(as, alarm)
				}
			}
			return as
		},
		"prefix": func(v interface{}) string {
			switch v.(type) {
			case *window.VPC:
				return "/vpc/" + v.(*window.VPC).Name
			case *window.Classic:
				return "/classic"
			case *window.Region:
				return ""
			}
			return ""
		},
		"split": func(vs []*window.CloudWatchAlarm) [][]*window.CloudWatchAlarm {
			var ret [][]*window.CloudWatchAlarm
			half := len(vs) / 2
			if len(vs) > half {
				ret = append(ret, vs[:half])
			}
			ret = append(ret, vs[half:])
			return ret
		},
		"trimFloat": func(v float64) string {
			return strconv.FormatFloat(v, 'f', -1, 64)
		},
		"rps": func(n float64, label string) string {
			if len(label) > 0 {
				label = " " + label
			}
			if n == 0 {
				return "0" + label
			}
			if n < 1 {
				n *= 60
				if n < 1 {
					if n < .1 {
						return "~0" + label + "/min"
					}
					return commify(strconv.FormatFloat(n, 'f', 1, 64)) + label + "/min"
				}
				return commify(strconv.FormatFloat(n, 'f', 0, 64)) + label + "/min"
			}
			return commify(strconv.FormatFloat(n, 'f', 0, 64)) + label + "/sec"
		},
		"commify": commify,
		"seconds": func(n int64) string {
			return (time.Duration(n) * time.Second).String()
		},
	}
	sizeNames []string = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
)

func commify(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	var num, frac string
	if i := strings.IndexByte(s, '.'); i > -1 {
		num, frac = s[:i], s[i:]
	} else {
		num = s
	}
	buf := make([]byte, 0, (2 * len(num)))
	if strings.HasPrefix(num, "-") {
		buf = append(buf, '-')
		num = num[1:]
	}
	if pn := len(num) % 3; pn > 0 {
		buf = append(buf, num[:pn]...)
		buf = append(buf, ',')
		num = num[pn:]
	}
	for len(num) > 0 {
		buf = append(buf, num[0], num[1], num[2], ',')
		num = num[3:]
	}
	if buf[len(buf)-1] == ',' {
		buf = buf[:len(buf)-1]
	}
	buf = append(buf, frac...)
	return string(buf)
}

func NewTemplateSet() *TemplateSet {
	ts := &TemplateSet{
		me: sync.RWMutex{},
	}
	if err := ts.Build(); err != nil {
		log.Println(err)
	}
	return ts
}

func (ts *TemplateSet) Build() error {
	ts.me.Lock()
	defer ts.me.Unlock()
	templates, err := template.New("").Funcs(templateFuncs).ParseFiles(templateFiles...)
	if err == nil {
		ts.templates = templates
	}
	return err
}

func (ts *TemplateSet) Execute(name string, data interface{}) string {
	ts.me.RLock()
	defer ts.me.RUnlock()
	if ts == nil {
		return "templates is nil"
	}
	var buf bytes.Buffer
	if err := ts.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return err.Error()
	}
	return buf.String()
}

func soft_true(v interface{}) bool {

	val := reflect.ValueOf(v)

	// resolve the pointer if it's valid
	for val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	if !val.IsValid() {
		return false
	}

	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() > 0
	case reflect.Bool:
		return val.Bool()
	case reflect.Complex64, reflect.Complex128:
		return val.Complex() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		return !val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0
	case reflect.Float32, reflect.Float64:
		return val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() != 0
	case reflect.Struct:
		return true // Struct values are always true.
	}

	return false

}

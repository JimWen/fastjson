// port from https://github.com/oliveagle/jsonpath
package fastjson

import (
	"errors"
	"fmt"
	"github.com/JimWen/fastjson/fastfloat"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"
)

var ErrGetFromNullObj = errors.New("get attribute from null object")

func JsonPathLookupRaw(obj *Value, jpath string) (interface{}, error) {
	c, err := Compile(jpath)
	if err != nil {
		return nil, err
	}

	v, err := c.Lookup(obj)
	if err != nil {
		return nil, err
	}

	switch v.t {
	case TypeObject:
		return v.o, nil
	case TypeArray:
		return v.a, nil
	case TypeString:
		return v.s, nil
	case typeRawString:
		return v.s, nil
	case TypeNumber:
		return fastfloat.ParseBestEffort(v.s), nil
	case TypeTrue:
		return true, nil
	case TypeFalse:
		return false, nil
	case TypeNull:
		return nil, nil
	}

	return nil, nil
}

func JsonPathLookup(obj *Value, jpath string) (*Value, error) {
	c, err := Compile(jpath)
	if err != nil {
		return nil, err
	}
	return c.Lookup(obj)
}

func JsonPathExists(obj *Value, jpath string) bool {
	c, err := Compile(jpath)
	if err != nil {
		return false
	}

	return c.Exists(obj)
}

type Compiled struct {
	path  string
	steps []step
}

type step struct {
	op   string
	key  string
	args interface{}
}

func MustCompile(jpath string) *Compiled {
	c, err := Compile(jpath)
	if err != nil {
		panic(err)
	}
	return c
}

func Compile(jpath string) (*Compiled, error) {
	tokens, err := tokenize(jpath)
	if err != nil {
		return nil, err
	}
	if tokens[0] != "@" && tokens[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of path")
	}
	tokens = tokens[1:]
	res := Compiled{
		path:  jpath,
		steps: make([]step, len(tokens)),
	}
	for i, token := range tokens {
		op, key, args, err := parse_token(token)
		if err != nil {
			return nil, err
		}
		res.steps[i] = step{op, key, args}
	}
	return &res, nil
}

func (c *Compiled) String() string {
	return fmt.Sprintf("Compiled lookup: %s", c.path)
}

func (c *Compiled) Exists(root *Value) bool {
	v, err := c.Lookup(root)
	if err != nil {
		return false
	}

	return v.Exists() && !v.Empty()
}

func (c *Compiled) Lookup(root *Value) (*Value, error) {
	var err error

	obj := root
	for _, s := range c.steps {
		// "key", "idx"
		switch s.op {
		case "key":
			// 支持$.key1.key2
			if len(s.key) > 0 {
				obj, err = get_key(obj, s.key)
				if err != nil {
					return nil, err
				}
			}

			// 如下支持 ['key1', 'key2'] or key3['key1', 'key2'] 计算
			if s.args == nil {
				continue
			}

			if len(s.args.([]string)) > 1 {
				res := &Value{a: make([]*Value, 0), t: TypeArray}
				for _, x := range s.args.([]string) {
					tmp, err := get_key(obj, x)
					if err != nil {
						return nil, err
					}
					res.a = append(res.a, tmp)
				}
				obj = res
			} else if len(s.args.([]string)) == 1 {
				obj, err = get_key(obj, s.args.([]string)[0])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("cannot index on empty key slice")
			}
		case "idx":
			if len(s.key) > 0 {
				// no key `$[0].test`
				obj, err = get_key(obj, s.key)
				if err != nil {
					return nil, err
				}
			}

			if len(s.args.([]int)) > 1 {
				res := &Value{a: make([]*Value, 0), t: TypeArray}
				for _, x := range s.args.([]int) {
					//fmt.Println("idx ---- ", x)
					tmp, err := get_idx(obj, x)
					if err != nil {
						return nil, err
					}
					res.a = append(res.a, tmp)
				}
				obj = res
			} else if len(s.args.([]int)) == 1 {
				//fmt.Println("idx ----------------3")
				obj, err = get_idx(obj, s.args.([]int)[0])
				if err != nil {
					return nil, err
				}
			} else {
				//fmt.Println("idx ----------------4")
				return nil, fmt.Errorf("cannot index on empty index slice")
			}
		case "range":
			if len(s.key) > 0 {
				// no key `$[:1].test`
				obj, err = get_key(obj, s.key)
				if err != nil {
					return nil, err
				}
			}
			if argsv, ok := s.args.([2]interface{}); ok == true {
				obj, err = get_range(obj, argsv[0], argsv[1])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("range args length should be 2")
			}
		case "filter":
			obj, err = get_key(obj, s.key)
			if err != nil {
				return nil, err
			}

			isArr, ret, err := get_filtered(obj, root, s.args.(string))
			if err != nil {
				return nil, err
			}

			if isArr {
				obj = &Value{a: ret, t: TypeArray}
			} else {
				if len(ret) == 0 {
					return nil, nil
				}

				obj = ret[0]
			}
		default:
			return nil, fmt.Errorf("expression don't support in filter")
		}
	}
	return obj, nil
}

func tokenize(query string) ([]string, error) {
	tokens := []string{}
	//	token_start := false
	//	token_end := false
	token := ""

	// fmt.Println("-------------------------------------------------- start")
	for idx, x := range query {
		token += string(x)
		// //fmt.Printf("idx: %d, x: %s, token: %s, tokens: %v\n", idx, string(x), token, tokens)
		if idx == 0 {
			if token == "$" || token == "@" {
				tokens = append(tokens, token[:])
				token = ""
				continue
			} else {
				return nil, fmt.Errorf("should start with '$'")
			}
		}
		if token == "." {
			continue
		} else if token == ".." {
			if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, "*")
			}
			token = "."
			continue
		} else {
			// fmt.Println("else: ", string(x), token)
			if strings.Contains(token, "[") {
				// fmt.Println(" contains [ ")
				if x == ']' && !strings.HasSuffix(token, "\\]") {
					if token[0] == '.' {
						tokens = append(tokens, token[1:])
					} else {
						tokens = append(tokens, token[:])
					}
					token = ""
					continue
				}
			} else {
				// fmt.Println(" doesn't contains [ ")
				if x == '.' {
					if token[0] == '.' {
						tokens = append(tokens, token[1:len(token)-1])
					} else {
						tokens = append(tokens, token[:len(token)-1])
					}
					token = "."
					continue
				}
			}
		}
	}
	if len(token) > 0 {
		if token[0] == '.' {
			token = token[1:]
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		} else {
			if token != "*" {
				tokens = append(tokens, token[:])
			} else if tokens[len(tokens)-1] != "*" {
				tokens = append(tokens, token[:])
			}
		}
	}
	// fmt.Println("finished tokens: ", tokens)
	// fmt.Println("================================================= done ")
	return tokens, nil
}

/*
op: "root", "key", "idx", "range", "filter", "scan"
*/
func parse_token(token string) (op string, key string, args interface{}, err error) {
	if token == "$" {
		return "root", "$", nil, nil
	}
	if token == "*" {
		return "scan", "*", nil, nil
	}

	bracket_idx := strings.Index(token, "[")
	if bracket_idx < 0 {
		return "key", token, nil, nil
	} else {
		key = token[:bracket_idx]
		tail := token[bracket_idx:]
		if len(tail) < 3 {
			err = fmt.Errorf("len(tail) should >=3, %v", tail)
			return
		}
		tail = tail[1 : len(tail)-1]

		//fmt.Println(key, tail)
		if strings.Contains(tail, "?") {
			// filter -------------------------------------------------
			op = "filter"
			if strings.HasPrefix(tail, "?(") && strings.HasSuffix(tail, ")") {
				args = strings.Trim(tail[2:len(tail)-1], " ")
			}
			return
		} else if strings.Contains(tail, ":") {
			// range ----------------------------------------------
			op = "range"
			tails := strings.Split(tail, ":")
			if len(tails) != 2 {
				err = fmt.Errorf("only support one range(from, to): %v", tails)
				return
			}
			var frm interface{}
			var to interface{}
			if frm, err = strconv.Atoi(strings.Trim(tails[0], " ")); err != nil {
				if strings.Trim(tails[0], " ") == "" {
					err = nil
				}
				frm = nil
			}
			if to, err = strconv.Atoi(strings.Trim(tails[1], " ")); err != nil {
				if strings.Trim(tails[1], " ") == "" {
					err = nil
				}
				to = nil
			}
			args = [2]interface{}{frm, to}
			return
		} else if tail == "*" {
			op = "range"
			args = [2]interface{}{nil, nil}
			return
		} else {
			op, args, err = parse_bracket_token(tail)
		}
	}

	return op, key, args, err
}

// 对jsonpath filter/idx/key  进一步解析
func parse_bracket_token(token string) (op string, args interface{}, err error) {

	hasKey := false
	hasOp := false
	hasLogic := false
	for _, v := range token {
		switch v {
		// 字符串引号
		case '\'':
			hasKey = true

		// 运算符运算符==  !=  <  <=  >  >=  =~
		case '=', '!', '<', '>', '~':
			hasOp = true

		// 逻辑运算符&&  ||
		case '&', '|':
			hasLogic = true

		default:
			// 判断除了数字 逗号 26字母全部不合符
			if (v < '0' || v > '9') && (v < 'a' || v > 'z') && (v < 'A' && v > 'Z') && v != ',' && v != '@' && v != '$' && v != '.' {
				return "", "", fmt.Errorf("invalid token: %v", token)
			}
		}
	}

	// 用运算符==  !=  <  <=  >  >=  =~和逻辑运算符&&  ||来连接的表达式 => filter
	if hasLogic || hasOp {
		op = "filter"
		args = strings.Trim(token, " ")
		return op, args, nil
	}

	// 1.全部是逗号分割的字符串，类似'key1','key2','key3' => key
	// 2.全部是字符串，类似'key' => key
	if hasKey {
		strs := strings.Split(token, ",")
		strArr := make([]string, len(strs))
		for i, str := range strs {
			strArr[i] = strings.Trim(strings.Trim(str, " "), "'")
		}

		op = "key"
		args = strArr

		return op, args, nil
	}

	// 1.全部是逗号分割的数字，类似0,1,2,3 => idx
	// 2.全部是数字，类似123 => idx
	res := []int{}

	for _, x := range strings.Split(token, ",") {
		if i, err := strconv.Atoi(strings.Trim(x, " ")); err == nil {
			res = append(res, i)
		} else {
			return "", nil, err
		}
	}

	op = "idx"
	args = res

	return op, args, nil
}

func filter_get_from_explicit_path(obj *Value, path string) (*Value, error) {
	steps, err := tokenize(path)
	//fmt.Println("f: steps: ", steps, err)
	//fmt.Println(path, steps)
	if err != nil {
		return nil, err
	}
	if steps[0] != "@" && steps[0] != "$" {
		return nil, fmt.Errorf("$ or @ should in front of filter path")
	}
	steps = steps[1:]
	xobj := obj
	//fmt.Println("f: xobj", xobj)
	for _, s := range steps {
		op, key, args, err := parse_token(s)
		// "key", "idx"
		switch op {
		case "key":
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}

		case "idx":
			if len(args.([]int)) != 1 {
				return nil, fmt.Errorf("don't support multiple index in filter")
			}
			xobj, err = get_key(xobj, key)
			if err != nil {
				return nil, err
			}
			xobj, err = get_idx(xobj, args.([]int)[0])
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("expression don't support in filter")
		}
	}

	return xobj, nil
}

func get_key(obj *Value, key string) (*Value, error) {
	if obj == nil {
		return nil, ErrGetFromNullObj
	}

	// 部分jsonpath表达式，key为空比如 $[@.ext == 2]，此时直接操作$对应的obj
	if len(key) == 0 {
		return obj, nil
	}

	switch obj.t {
	case TypeObject:
		// if obj came from stdlib json, its highly likely to be a map[string]interface{}
		// in which case we can save having to iterate the map keys to work out if the
		// key exists
		if !obj.Exists(key) {
			return nil, fmt.Errorf("key error: %s not found in object", key)
		}

		return obj.Get(key), nil

	case TypeArray:
		// slice we should get from all objects in it.
		res := &Value{a: make([]*Value, 0), t: TypeArray}
		for _, v := range obj.a {
			res.a = append(res.a, v.Get(key))
		}
		return res, nil

	default:
		return nil, fmt.Errorf("fail to exec get_key:%s, object is not map", key)
	}
}

func get_idx(obj *Value, idx int) (*Value, error) {
	switch obj.t {
	case TypeArray:
		length := len(obj.a)

		if idx >= 0 {
			if idx >= length {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return obj.a[idx], nil
		} else {
			// < 0
			_idx := length + idx
			if _idx < 0 {
				return nil, fmt.Errorf("index out of range: len: %v, idx: %v", length, idx)
			}
			return obj.a[_idx], nil
		}

	default:
		return nil, fmt.Errorf("fail to exec get_idx:%d, object is not Slice", idx)
	}
}

func get_range(obj *Value, frm, to interface{}) (*Value, error) {
	switch obj.t {
	case TypeArray:
		length := len(obj.a)

		_frm := 0
		_to := length
		if frm == nil {
			frm = 0
		}
		if to == nil {
			to = length - 1
		}

		if fv, ok := frm.(int); ok == true {
			if fv < 0 {
				_frm = length + fv
			} else {
				_frm = fv
			}
		}

		if tv, ok := to.(int); ok == true {
			if tv < 0 {
				_to = length + tv + 1
			} else {
				_to = tv + 1
			}
		}

		if _frm < 0 || _frm >= length {
			return nil, fmt.Errorf("index [from] out of range: len: %v, from: %v", length, frm)
		}
		if _to < 0 || _to > length {
			return nil, fmt.Errorf("index [to] out of range: len: %v, to: %v", length, to)
		}

		//fmt.Println("_frm, _to: ", _frm, _to)
		return &Value{a: obj.a[_frm:_to], t: TypeArray}, nil

	default:
		return nil, fmt.Errorf("fail to exec get_idx:from %v to %v, object is not Slice", frm, to)
	}
}

func regFilterCompile(rule string) (*regexp.Regexp, error) {
	runes := []rune(rule)
	if len(runes) <= 2 {
		return nil, errors.New("empty rule")
	}

	if runes[0] != '/' || runes[len(runes)-1] != '/' {
		return nil, errors.New("invalid syntax. should be in `/pattern/` form")
	}
	runes = runes[1 : len(runes)-1]
	return regexp.Compile(string(runes))
}

func get_filtered(obj, root *Value, filter string) (bool, []*Value, error) {
	lp, op, rp, err := parse_filter(filter)
	if err != nil {
		return true, nil, err
	}

	res := make([]*Value, 0)

	switch obj.t {
	case TypeArray:
		if op == "=~" {
			// regexp
			pat, err := regFilterCompile(rp)
			if err != nil {
				return true, nil, err
			}

			for _, tmp := range obj.a {
				ok, err := eval_reg_filter(tmp, root, lp, pat)
				if err != nil {
					return true, nil, err
				}
				if ok == true {
					res = append(res, tmp)
				}
			}
		} else {
			for _, tmp := range obj.a {
				ok, err := eval_filter(tmp, root, lp, op, rp)
				if err != nil {
					return true, nil, err
				}
				if ok == true {
					res = append(res, tmp)
				}
			}
		}

		return true, res, nil

	case TypeObject:
		if op == "=~" {
			// regexp
			pat, err := regFilterCompile(rp)
			if err != nil {
				return false, nil, err
			}

			ok, err := eval_reg_filter(obj, root, lp, pat)
			if err != nil {
				return false, nil, err
			}
			if ok == true {
				res = append(res, obj)
			}
		} else {
			ok, err := eval_filter(obj, root, lp, op, rp)
			if err != nil {
				return false, nil, err
			}
			if ok == true {
				res = append(res, obj)
			}
		}

		return false, res, nil

	default:
		return true, nil, fmt.Errorf("don't support filter on this type: %v", obj.t)
	}
}

// @.isbn                 => @.isbn, exists, nil
// @.price < 10           => @.price, <, 10
// @.price <= $.expensive => @.price, <=, $.expensive
// @.author =~ /.*REES/i  => @.author, match, /.*REES/i

func parse_filter(filter string) (lp string, op string, rp string, err error) {
	tmp := ""

	stage := 0
	str_embrace := false
	for idx, c := range filter {
		switch c {
		case '\'':
			if str_embrace == false {
				str_embrace = true
			} else {
				switch stage {
				case 0:
					lp = tmp
				case 1:
					op = tmp
				case 2:
					rp = tmp
				}
				tmp = ""
			}
		case ' ':
			if str_embrace == true {
				tmp += string(c)
				continue
			}
			switch stage {
			case 0:
				lp = tmp
			case 1:
				op = tmp
			case 2:
				rp = tmp
			}
			tmp = ""

			stage += 1
			if stage > 2 {
				return "", "", "", errors.New(fmt.Sprintf("invalid char at %d: `%c`", idx, c))
			}
		default:
			tmp += string(c)
		}
	}
	if tmp != "" {
		switch stage {
		case 0:
			lp = tmp
			op = "exists"
		case 1:
			op = tmp
		case 2:
			rp = tmp
		}
		tmp = ""
	}
	return lp, op, rp, err
}

func parse_filter_v1(filter string) (lp string, op string, rp string, err error) {
	tmp := ""
	istoken := false
	for _, c := range filter {
		if istoken == false && c != ' ' {
			istoken = true
		}
		if istoken == true && c == ' ' {
			istoken = false
		}
		if istoken == true {
			tmp += string(c)
		}
		if istoken == false && tmp != "" {
			if lp == "" {
				lp = tmp[:]
				tmp = ""
			} else if op == "" {
				op = tmp[:]
				tmp = ""
			} else if rp == "" {
				rp = tmp[:]
				tmp = ""
			}
		}
	}
	if tmp != "" && lp == "" && op == "" && rp == "" {
		lp = tmp[:]
		op = "exists"
		rp = ""
		err = nil
		return
	} else if tmp != "" && rp == "" {
		rp = tmp[:]
		tmp = ""
	}
	return lp, op, rp, err
}

func eval_reg_filter(obj, root *Value, lp string, pat *regexp.Regexp) (res bool, err error) {
	if pat == nil {
		return false, errors.New("nil pat")
	}
	lp_v, err := get_lp_v(obj, root, lp)
	if err != nil {
		return false, err
	}
	switch lp_v.Type() {
	case TypeString:
		return pat.MatchString(lp_v.s), nil
	default:
		return false, errors.New("only string can match with regular expression")
	}
}

func get_lp_v(obj, root *Value, lp string) (*Value, error) {
	var lp_v *Value
	if strings.HasPrefix(lp, "@.") {
		return filter_get_from_explicit_path(obj, lp)
	} else if strings.HasPrefix(lp, "$.") {
		return filter_get_from_explicit_path(root, lp)
	} else {
		lp_v = &Value{s: lp, t: TypeString}
	}
	return lp_v, nil
}

func eval_filter(obj, root *Value, lp, op, rp string) (res bool, err error) {
	lp_v, err := get_lp_v(obj, root, lp)

	if op == "exists" {
		return lp_v != nil, nil
	} else if op == "=~" {
		return false, fmt.Errorf("not implemented yet")
	} else {
		var rp_v *Value
		if strings.HasPrefix(rp, "@.") {
			rp_v, err = filter_get_from_explicit_path(obj, rp)
		} else if strings.HasPrefix(rp, "$.") {
			rp_v, err = filter_get_from_explicit_path(root, rp)
		} else {
			rp_v = &Value{s: rp, t: TypeString}
		}

		//fmt.Printf("lp_v: %v, rp_v: %v\n", lp_v, rp_v)
		return cmp_any(lp_v, rp_v, op)
	}
}

func isNumber(o *Value) bool {
	switch o.Type() {
	case TypeNumber:
		return true
	case TypeString:
		_, err := strconv.ParseFloat(o.s, 64)
		if err == nil {
			return true
		} else {
			return false
		}
	}
	return false
}

func cmp_any(obj1, obj2 *Value, op string) (bool, error) {
	switch op {
	case "<", "<=", "==", "!=", ">=", ">":
	default:
		return false, fmt.Errorf("op should only be <, <=, ==, >= and >")
	}

	var exp string
	if isNumber(obj1) && isNumber(obj2) {
		exp = fmt.Sprintf(`%v %s %v`, obj1.s, op, obj2.s)
	} else {
		exp = fmt.Sprintf(`"%v" %s "%v"`, obj1.s, op, obj2.s)
	}

	//fmt.Println("exp: ", exp)
	fset := token.NewFileSet()
	res, err := types.Eval(fset, nil, 0, exp)
	if err != nil {
		return false, err
	}

	if res.IsValue() == false || (res.Value.String() != "false" && res.Value.String() != "true") {
		return false, fmt.Errorf("result should only be true or false")
	}
	if res.Value.String() == "true" {
		return true, nil
	}

	return false, nil
}

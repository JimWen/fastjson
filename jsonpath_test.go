package fastjson

import (
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"regexp"
	"testing"
)

var json_data *Value

func init() {
	data := `
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10,
	"expensive2": 100,
	"expensive3": 1000
}
`

	json_data, _ = Parse(data)
}

func Test_jsonpath_JsonPathLookup_1(t *testing.T) {
	// key from root
	res, _ := JsonPathLookup(json_data, "$.expensive")
	if res_v, ok := res.ToFloat64(); ok != nil || res_v != 10.0 {
		t.Errorf("expensive should be 10")
	}

	// single index
	res, _ = JsonPathLookup(json_data, "$.store.book[0].price")
	if res_v, ok := res.ToFloat64(); ok != nil || res_v != 8.95 {
		t.Errorf("$.store.book[0].price should be 8.95")
	}

	// nagtive single index
	res, _ = JsonPathLookup(json_data, "$.store.book[-1].isbn")
	if res_v, ok := res.ToString(); ok != nil || res_v != "0-395-19395-8" {
		t.Errorf("$.store.book[-1].isbn should be \"0-395-19395-8\" %s", res_v)
	}

	// multiple index
	res, err := JsonPathLookup(json_data, "$.store.book[0,1].price")
	t.Log(err, res)
	if res_v, ok := res.ToArray(); ok != nil || res_v[0].MustFloat64() != 8.95 || res_v[1].MustFloat64() != 12.99 {
		t.Errorf("exp: [8.95, 12.99], got: %v", res)
	}

	// multiple index
	res, err = JsonPathLookup(json_data, "$.store.book[0,1].title")
	t.Log(err, res)
	if res_v, ok := res.ToArray(); ok != nil {
		if res_v[0].MustString() != "Sayings of the Century" || res_v[1].MustString() != "Sword of Honour" {
			t.Errorf("title are wrong: %v", res)
		}
	}

	// full array
	res, err = JsonPathLookup(json_data, "$.store.book[0:].price")
	t.Log(err, res)
	if res_v, ok := res.ToArray(); ok != nil || res_v[0].MustFloat64() != 8.95 || res_v[1].MustFloat64() != 12.99 || res_v[2].MustFloat64() != 8.99 || res_v[3].MustFloat64() != 22.99 {
		t.Errorf("exp: [8.95, 12.99, 8.99, 22.99], got: %v", res)
	}

	// range
	res, err = JsonPathLookup(json_data, "$.store.book[0:1].price")
	t.Log(err, res)
	if res_v, ok := res.ToArray(); ok != nil || res_v[0].MustFloat64() != 8.95 || res_v[1].MustFloat64() != 12.99 {
		t.Errorf("exp: [8.95, 12.99], got: %v", res)
	}

	// range
	res, err = JsonPathLookup(json_data, "$.store.book[0:1].title")
	t.Log(err, res)
	if res_v, ok := res.ToArray(); ok != nil {
		if res_v[0].MustString() != "Sayings of the Century" || res_v[1].MustString() != "Sword of Honour" {
			t.Errorf("title are wrong: %v", res)
		}
	}
}

func Test_jsonpath_JsonPathExists_1(t *testing.T) {
	// key from root
	res := JsonPathExists(json_data, "$.expensive")
	t.Log(res)

	res = JsonPathExists(json_data, "$.expensive2")
	t.Log(res)

	// single index
	res = JsonPathExists(json_data, "$.store.book[0].price")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[0].price2")
	t.Log(res)

	// nagtive single index
	res = JsonPathExists(json_data, "$.store.book[-1].isbn")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[-1].isbn2")
	t.Log(res)

	// multiple index
	res = JsonPathExists(json_data, "$.store.book[0,1].price")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[0,1].price2")
	t.Log(res)

	// multiple index
	res = JsonPathExists(json_data, "$.store.book[0,1].title")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[0,1].title3")
	t.Log(res)

	// full array
	res = JsonPathExists(json_data, "$.store.book[0:].price")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[0:].price2")
	t.Log(res)

	// judge
	res = JsonPathExists(json_data, "$.store.book[?(@.isbn)].isbn")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[?(@.isbn2)].isbn")
	t.Log(res)

	// compare
	res = JsonPathExists(json_data, "$.store.book[?(@.price > 10)].title")
	t.Log(res)

	res = JsonPathExists(json_data, "$.store.book[?(@.price > 10)].title2")
	t.Log(res)
}

func Test_jsonpath_JsonPathLookup_filter(t *testing.T) {
	res, err := JsonPathLookup(json_data, "$.store.book[?(@.isbn)].isbn")
	t.Log(err, res)

	if res_v, ok := res.ToArray(); ok != nil {
		if res_v[0].MustString() != "0-553-21311-3" || res_v[1].MustString() != "0-395-19395-8" {
			t.Errorf("error: %v", res)
		}
	}

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price > 10)].title")
	t.Log(err, res)
	if res_v, ok := res.ToArray(); ok != nil {
		if res_v[0].MustString() != "Sword of Honour" || res_v[1].MustString() != "The Lord of the Rings" {
			t.Errorf("error: %v", res)
		}
	}

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price > 10)]")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price > $.expensive)].price")
	t.Log(err, res)
	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price < $.expensive)].price")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$.store.book[?(@.price != 12.99)]")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$.expensive")
	t.Log(err, res)
}

func Test_jsonpath_authors_of_all_books(t *testing.T) {
	query := "store.book[*].author"
	expected := []string{
		"Nigel Rees",
		"Evelyn Waugh",
		"Herman Melville",
		"J. R. R. Tolkien",
	}
	res, _ := JsonPathLookup(json_data, query)
	t.Log(res, expected)
}

var token_cases = []map[string]interface{}{
	map[string]interface{}{
		"query":  "$..author",
		"tokens": []string{"$", "*", "author"},
	},
	map[string]interface{}{
		"query":  "$.store.*",
		"tokens": []string{"$", "store", "*"},
	},
	map[string]interface{}{
		"query":  "$.store..price",
		"tokens": []string{"$", "store", "*", "price"},
	},
	map[string]interface{}{
		"query":  "$.store.book[*].author",
		"tokens": []string{"$", "store", "book[*]", "author"},
	},
	map[string]interface{}{
		"query":  "$..book[2]",
		"tokens": []string{"$", "*", "book[2]"},
	},
	map[string]interface{}{
		"query":  "$..book[(@.length-1)]",
		"tokens": []string{"$", "*", "book[(@.length-1)]"},
	},
	map[string]interface{}{
		"query":  "$..book[0,1]",
		"tokens": []string{"$", "*", "book[0,1]"},
	},
	map[string]interface{}{
		"query":  "$..book[:2]",
		"tokens": []string{"$", "*", "book[:2]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.isbn)]",
		"tokens": []string{"$", "*", "book[?(@.isbn)]"},
	},
	map[string]interface{}{
		"query":  "$.store.book[?(@.price < 10)]",
		"tokens": []string{"$", "store", "book[?(@.price < 10)]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.price <= $.expensive)]",
		"tokens": []string{"$", "*", "book[?(@.price <= $.expensive)]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.author =~ /.*REES/i)]",
		"tokens": []string{"$", "*", "book[?(@.author =~ /.*REES/i)]"},
	},
	map[string]interface{}{
		"query":  "$..book[?(@.author =~ /.*REES\\]/i)]",
		"tokens": []string{"$", "*", "book[?(@.author =~ /.*REES\\]/i)]"},
	},
	map[string]interface{}{
		"query":  "$..*",
		"tokens": []string{"$", "*"},
	},
	map[string]interface{}{
		"query":  "$....author",
		"tokens": []string{"$", "*", "author"},
	},
}

func Test_jsonpath_tokenize(t *testing.T) {
	for idx, tcase := range token_cases {
		t.Logf("idx[%d], tcase: %v", idx, tcase)
		query := tcase["query"].(string)
		expected_tokens := tcase["tokens"].([]string)
		tokens, err := tokenize(query, false)
		t.Log(err, tokens, expected_tokens)
		if len(tokens) != len(expected_tokens) {
			t.Errorf("different length: (got)%v, (expected)%v", len(tokens), len(expected_tokens))
			continue
		}
		for i := 0; i < len(expected_tokens); i++ {
			if tokens[i] != expected_tokens[i] {
				t.Errorf("not expected: [%d], (got)%v != (expected)%v", i, tokens[i], expected_tokens[i])
			}
		}
	}
}

var parse_token_cases = []map[string]interface{}{

	map[string]interface{}{
		"token": "$",
		"op":    "root",
		"key":   "$",
		"args":  nil,
	},
	map[string]interface{}{
		"token": "store",
		"op":    "key",
		"key":   "store",
		"args":  nil,
	},

	// idx --------------------------------------
	map[string]interface{}{
		"token": "book[2]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{2},
	},
	map[string]interface{}{
		"token": "book[-1]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{-1},
	},
	map[string]interface{}{
		"token": "book[0,1]",
		"op":    "idx",
		"key":   "book",
		"args":  []int{0, 1},
	},
	map[string]interface{}{
		"token": "[0]",
		"op":    "idx",
		"key":   "",
		"args":  []int{0},
	},

	// range ------------------------------------
	map[string]interface{}{
		"token": "book[1:-1]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{1, -1},
	},
	map[string]interface{}{
		"token": "book[*]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{nil, nil},
	},
	map[string]interface{}{
		"token": "book[:2]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{nil, 2},
	},
	map[string]interface{}{
		"token": "book[-2:]",
		"op":    "range",
		"key":   "book",
		"args":  [2]interface{}{-2, nil},
	},

	// filter --------------------------------
	map[string]interface{}{
		"token": "book[?( @.isbn      )]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.isbn",
	},
	map[string]interface{}{
		"token": "book[?(@.price < 10)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.price < 10",
	},
	map[string]interface{}{
		"token": "book[?(@.price <= $.expensive)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.price <= $.expensive",
	},
	map[string]interface{}{
		"token": "book[?(@.author =~ /.*REES/i)]",
		"op":    "filter",
		"key":   "book",
		"args":  "@.author =~ /.*REES/i",
	},
	map[string]interface{}{
		"token": "*",
		"op":    "scan",
		"key":   "*",
		"args":  nil,
	},
}

func Test_my_jsonpath_parse_filter_group(t *testing.T) {
	token := "@.expensive == 10 || @.expensive == 9"

	ret, err := parse_filter_group(token)
	t.Log(err, ret)

	token = "@.price < 10 && @.author =~ /.*REES/"
	ret, err = parse_filter_group(token)
	t.Log(err, ret)

	token = "@.price < 10 || @.author"
	ret, err = parse_filter_group(token)
	t.Log(err, ret)

	token = "@.price < 10 || @.author == 'xxx'"
	ret, err = parse_filter_group(token)
	t.Log(err, ret)

	token = "@.price < 10 || @.author == 'xxx' && @.author != 'yyy'"
	ret, err = parse_filter_group(token)
	t.Log(err, ret)
}

func Test_my_jsonpath_tokenize_filter(t *testing.T) {
	token := "@.expensive == 10"
	ret, err := tokenize(token, true)
	t.Log(ret, err)

	token = "expensive == 10"
	ret, err = tokenize(token, true)
	t.Log(ret, err)

	token = "$[container_id != '']"

	op, key, args, err := parse_token(token)
	t.Logf("got: err: %v, op: %v, key: %v, args: %v\n", err, op, key, args)
}

func Test_my_jsonpath_parse_token(t *testing.T) {
	res, err := JsonPathLookup(json_data, "$[?(@.expensive == 10)].store.book")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[@.expensive == 10 || @.expensive == 9].store.book")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[expensive == 10].store.book")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[@.expensive == 9].store.book")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[?(@.expensive == 10)]['store']['book']")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[?(@.expensive == 10)]['store']['book']['category']")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[?(@.expensive == 10)]['expensive']")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[?(@.expensive == 10)]['expensive', 'expensive2']")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[expensive == 10]['expensive', 'expensive2']")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[@.expensive == 10]['store']['book']['category']")
	t.Log(err, res)

	res, err = JsonPathLookup(json_data, "$[@.expensive == 10]['store']['book']['category', 'author']")
	t.Log(err, res)
}

func Test_jsonpath_parse_token(t *testing.T) {
	for idx, tcase := range parse_token_cases {
		t.Logf("[%d] - tcase: %v", idx, tcase)
		token := tcase["token"].(string)
		exp_op := tcase["op"].(string)
		exp_key := tcase["key"].(string)
		exp_args := tcase["args"]

		op, key, args, err := parse_token(token)
		t.Logf("[%d] - expected: op: %v, key: %v, args: %v\n", idx, exp_op, exp_key, exp_args)
		t.Logf("[%d] - got: err: %v, op: %v, key: %v, args: %v\n", idx, err, op, key, args)
		if op != exp_op {
			t.Errorf("ERROR: op(%v) != exp_op(%v)", op, exp_op)
			return
		}
		if key != exp_key {
			t.Errorf("ERROR: key(%v) != exp_key(%v)", key, exp_key)
			return
		}

		if op == "idx" {
			if args_v, ok := args.([]int); ok == true {
				for i, v := range args_v {
					if v != exp_args.([]int)[i] {
						t.Errorf("ERROR: different args: [%d], (got)%v != (exp)%v", i, v, exp_args.([]int)[i])
						return
					}
				}
			} else {
				t.Errorf("ERROR: idx op should expect args:[]int{} in return, (got)%v", reflect.TypeOf(args))
				return
			}
		}

		if op == "range" {
			if args_v, ok := args.([2]interface{}); ok == true {
				fmt.Println(args_v)
				exp_from := exp_args.([2]interface{})[0]
				exp_to := exp_args.([2]interface{})[1]
				if args_v[0] != exp_from {
					t.Errorf("(from)%v != (exp_from)%v", args_v[0], exp_from)
					return
				}
				if args_v[1] != exp_to {
					t.Errorf("(to)%v != (exp_to)%v", args_v[1], exp_to)
					return
				}
			} else {
				t.Errorf("ERROR: range op should expect args:[2]interface{}, (got)%v", reflect.TypeOf(args))
				return
			}
		}

		if op == "filter" {
			if args_v, ok := args.(string); ok == true {
				fmt.Println(args_v)
				if exp_args.(string) != args_v {
					t.Errorf("len(args) not expected: (got)%v != (exp)%v", len(args_v), len(exp_args.([]string)))
					return
				}

			} else {
				t.Errorf("ERROR: filter op should expect args:[]string{}, (got)%v", reflect.TypeOf(args))
			}
		}
	}
}

func Test_jsonpath_get_key(t *testing.T) {
	obj, _ := Parse("{\"key\": 1}")

	res, err := get_key(obj, "key")
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get key: %v", err)
		return
	}
	if res.MustInt() != 1 {
		t.Errorf("key value is not 1: %v", res)
		return
	}

	res, err = get_key(obj, "hah")
	fmt.Println(err, res)
	if err == nil {
		t.Errorf("key error not raised")
		return
	}
	if res != nil {
		t.Errorf("key error should return nil res: %v", res)
		return
	}

	obj2 := &Value{s: "1", t: TypeNumber}
	res, err = get_key(obj2, "key")
	fmt.Println(err, res)
	if err == nil {

		t.Errorf("object is not map error not raised")
		return
	}

	obj3, _ := Parse("{\"key\": \"hah\"}")
	res, err = get_key(obj3, "key")
	if res_v, ok := res.ToString(); ok != nil || res_v != "hah" {
		fmt.Println(err, res)
		t.Errorf("map[string]string support failed")
	}

	obj4, _ := Parse("{{\"a\": 1}, {\"a\": 2}}")
	res, err = get_key(obj4, "a")
	fmt.Println(err, res)
}

func Test_jsonpath_get_idx(t *testing.T) {
	obj, _ := Parse("[1, 2, 3, 4]")

	res, err := get_idx(obj, 0)
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get_idx(obj,0): %v", err)
		return
	}
	if v, ok := res.ToInt(); ok != nil || v != 1 {
		t.Errorf("failed to get int 1")
	}

	res, err = get_idx(obj, 2)
	fmt.Println(err, res)
	if v, ok := res.ToInt(); ok != nil || v != 3 {
		t.Errorf("failed to get int 3")
	}
	res, err = get_idx(obj, 4)
	fmt.Println(err, res)
	if err == nil {
		t.Errorf("index out of range  error not raised")
		return
	}

	res, err = get_idx(obj, -1)
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get_idx(obj, -1): %v", err)
		return
	}
	if v, ok := res.ToInt(); ok != nil || v != 4 {
		t.Errorf("failed to get int 4")
	}

	res, err = get_idx(obj, -4)
	fmt.Println(err, res)
	if v, ok := res.ToInt(); ok != nil || v != 1 {
		t.Errorf("failed to get int 1")
	}

	res, err = get_idx(obj, -5)
	fmt.Println(err, res)
	if err == nil {
		t.Errorf("index out of range  error not raised")
		return
	}

	obj1 := &Value{s: "1", t: TypeNumber}
	res, err = get_idx(obj1, 1)
	if err == nil {
		t.Errorf("object is not Slice error not raised")
		return
	}

	obj2, _ := Parse("[1, 2, 3, 4]")
	res, err = get_idx(obj2, 0)
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get_idx(obj2,0): %v", err)
		return
	}
	if v, ok := res.ToInt(); ok != nil || v != 1 {
		t.Errorf("failed to get int 1")
	}
}

func Test_jsonpath_get_range(t *testing.T) {
	obj, _ := Parse("[1, 2, 3, 4, 5]")

	res, err := get_range(obj, 0, 2)
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get_range: %v", err)
	}

	ret, _ := res.ToArray()
	if ret[0].MustInt() != 1 || ret[1].MustInt() != 2 {
		t.Errorf("failed get_range: %v, expect: [1,2]", res)
	}

	obj1, _ := Parse("[1, 2, 3, 4, 5]")
	res, err = get_range(obj1, 3, -1)
	fmt.Println(err, res)
	if err != nil {
		t.Errorf("failed to get_range: %v", err)
	}

	ret, _ = res.ToArray()
	fmt.Println(ret)
	if ret[0].MustInt() != 4 || ret[1].MustInt() != 5 {
		t.Errorf("failed get_range: %v, expect: [4,5]", res)
	}

	res, err = get_range(obj1, nil, 2)
	t.Logf("err: %v, res:%v", err, res)
	ret, _ = res.ToArray()
	if ret[0].MustInt() != 1 || ret[1].MustInt() != 2 {
		t.Errorf("from support nil failed: %v", res)
	}

	res, err = get_range(obj1, nil, nil)
	t.Logf("err: %v, res:%v", err, res)
	ret, _ = res.ToArray()
	if len(ret) != 5 {
		t.Errorf("from, to both nil failed")
	}

	res, err = get_range(obj1, -2, nil)
	t.Logf("err: %v, res:%v", err, res)
	ret, _ = res.ToArray()
	if ret[0].MustInt() != 4 || ret[1].MustInt() != 5 {
		t.Errorf("from support nil failed: %v", res)
	}

	obj2 := &Value{s: "2", t: TypeNumber}
	res, err = get_range(obj2, 0, 1)
	fmt.Println(err, res)
	if err == nil {
		t.Errorf("object is Slice error not raised")
	}
}

func Test_jsonpath_types_eval(t *testing.T) {
	fset := token.NewFileSet()
	res, err := types.Eval(fset, nil, 0, "1 < 2")
	fmt.Println(err, res, res.Type, res.Value, res.IsValue())
}

var tcase_parse_filter = []map[string]interface{}{
	// 0
	map[string]interface{}{
		"filter":  "@.isbn",
		"exp_lp":  "@.isbn",
		"exp_op":  "exists",
		"exp_rp":  "",
		"exp_err": nil,
	},
	// 1
	map[string]interface{}{
		"filter":  "@.price < 10",
		"exp_lp":  "@.price",
		"exp_op":  "<",
		"exp_rp":  "10",
		"exp_err": nil,
	},
	// 2
	map[string]interface{}{
		"filter":  "@.price <= $.expensive",
		"exp_lp":  "@.price",
		"exp_op":  "<=",
		"exp_rp":  "$.expensive",
		"exp_err": nil,
	},
	// 3
	map[string]interface{}{
		"filter":  "@.author =~ /.*REES/i",
		"exp_lp":  "@.author",
		"exp_op":  "=~",
		"exp_rp":  "/.*REES/i",
		"exp_err": nil,
	},

	// 4
	{
		"filter": "@.author == 'Nigel Rees'",
		"exp_lp": "@.author",
		"exp_op": "==",
		"exp_rp": "Nigel Rees",
	},
}

func Test_jsonpath_parse_filter(t *testing.T) {

	//for _, tcase := range tcase_parse_filter[4:] {
	for _, tcase := range tcase_parse_filter {
		ret, _ := parse_filter(tcase["filter"].(string))
		t.Log(tcase)
		t.Logf("lp: %v, op: %v, rp: %v", ret.lp.p, ret.op, ret.rp.p)
		if ret.lp.p != tcase["exp_lp"].(string) {
			t.Errorf("%s(got) != %v(exp_lp)", ret.lp.p, tcase["exp_lp"])
			return
		}
		if ret.op != tcase["exp_op"].(string) {
			t.Errorf("%s(got) != %v(exp_op)", ret.op, tcase["exp_op"])
			return
		}
		if ret.rp.p != tcase["exp_rp"].(string) {
			t.Errorf("%s(got) != %v(exp_rp)", ret.rp.p, tcase["exp_rp"])
			return
		}
	}
}

var tcase_filter_get_from_explicit_path = []map[string]interface{}{
	// 0
	map[string]interface{}{
		// 0 {"a": 1}
		"obj":      MustParse("{\"a\": 1}"),
		"query":    "$.a",
		"expected": 1,
	},
	map[string]interface{}{
		// 1 {"a":{"b":1}}
		"obj":      MustParse("{\"a\": {\"b\": 1}}"),
		"query":    "$.a.b",
		"expected": 1,
	},
	map[string]interface{}{
		// 2 {"a": {"b":1, "c":2}}
		"obj":      MustParse("{\"a\": {\"b\": 1, \"c\": 2}}"),
		"query":    "$.a.c",
		"expected": 2,
	},
	map[string]interface{}{
		// 3 {"a": {"b":1}, "b": 2}
		"obj":      MustParse("{\"a\": {\"b\": 1}, \"b\": 2}"),
		"query":    "$.a.b",
		"expected": 1,
	},
	map[string]interface{}{
		// 4 {"a": {"b":1}, "b": 2}
		"obj":      MustParse("{\"a\": {\"b\": 1}, \"b\": 2}"),
		"query":    "$.b",
		"expected": 2,
	},
	map[string]interface{}{
		// 5 {'a': ['b',1]}
		"obj":      MustParse("{\"a\": [\"b\",1]}"),
		"query":    "$.a[0]",
		"expected": "b",
	},
}

func Test_jsonpath_filter_get_from_explicit_path(t *testing.T) {

	for idx, tcase := range tcase_filter_get_from_explicit_path {
		obj := tcase["obj"].(*Value)
		query := tcase["query"].(string)
		expected := tcase["expected"]

		res, err := filter_get_from_explicit_path(obj, query)
		t.Log(idx, err, res)
		if err != nil {
			t.Errorf("flatten_cases: failed: [%d] %v", idx, err)
		}

		t.Logf("typeof(res): %v, typeof(expected): %v", res.t, reflect.TypeOf(expected))
	}
}

var tcase_eval_filter = []map[string]interface{}{
	// 0
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 1}"),
		"root": MustParse("{}"),
		"lp":   "@.a",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 1
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 1}"),
		"root": MustParse("{}"),
		"lp":   "@.b",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},
	// 2
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 1}"),
		"root": MustParse("{\"a\": 1}"),
		"lp":   "$.a",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 3
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 1}"),
		"root": MustParse("{\"a\": 1}"),
		"lp":   "$.b",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},
	// 4
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 1, \"b\": {\"c\": 2}}"),
		"root": MustParse("{\"a\": 1, \"b\": {\"c\": 2}}"),
		"lp":   "$.b.c",
		"op":   "exists",
		"rp":   "",
		"exp":  true,
	},
	// 5
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 1, \"b\": {\"c\": 2}}"),
		"root": MustParse("{}"),
		"lp":   "$.b.a",
		"op":   "exists",
		"rp":   "",
		"exp":  false,
	},

	// 6
	map[string]interface{}{
		"obj":  MustParse("{\"a\": 3}"),
		"root": MustParse("{\"a\": 3}"),
		"lp":   "$.a",
		"op":   ">",
		"rp":   "1",
		"exp":  true,
	},
}

func Test_jsonpath_eval_filter(t *testing.T) {
	for idx, tcase := range tcase_eval_filter[1:] {
		fmt.Println("------------------------------")
		obj := tcase["obj"].(*Value)
		root := tcase["root"].(*Value)
		lp := tcase["lp"].(string)
		op := tcase["op"].(string)
		rp := tcase["rp"].(string)
		exp := tcase["exp"].(bool)
		t.Logf("idx: %v, lp: %v, op: %v, rp: %v, exp: %v", idx, lp, op, rp, exp)
		got, err := eval_filter(obj, root, FilterTuple{lp: Param{p: lp}, op: op, rp: Param{p: rp}})

		if err != nil {
			t.Errorf("idx: %v, failed to eval: %v", idx, err)
			return
		}
		if got != exp {
			t.Errorf("idx: %v, %v(got) != %v(exp)", idx, got, exp)
		}

	}
}

var (
	ifc1 *Value = &Value{s: "haha", t: TypeString}
	ifc2 *Value = &Value{s: "ha ha", t: TypeString}
)
var tcase_cmp_any = []map[string]interface{}{

	map[string]interface{}{
		"obj1": &Value{s: "1", t: TypeNumber},
		"obj2": &Value{s: "1", t: TypeNumber},
		"op":   "==",
		"exp":  true,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": &Value{s: "1", t: TypeNumber},
		"obj2": &Value{s: "2", t: TypeNumber},
		"op":   "==",
		"exp":  false,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": &Value{s: "1.0", t: TypeNumber},
		"obj2": &Value{s: "2.2", t: TypeNumber},
		"op":   "<",
		"exp":  true,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": &Value{s: "1", t: TypeNumber},
		"obj2": &Value{s: "2.0", t: TypeNumber},
		"op":   "<",
		"exp":  true,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": &Value{s: "1", t: TypeNumber},
		"obj2": &Value{s: "2.0", t: TypeNumber},
		"op":   ">",
		"exp":  false,
		"err":  nil,
	},
	map[string]interface{}{
		"obj1": &Value{s: "1", t: TypeNumber},
		"obj2": &Value{s: "2", t: TypeNumber},
		"op":   "=~",
		"exp":  false,
		"err":  "op should only be <, <=, ==, >= and >",
	}, {
		"obj1": ifc1,
		"obj2": ifc1,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	}, {
		"obj1": ifc2,
		"obj2": ifc2,
		"op":   "==",
		"exp":  true,
		"err":  nil,
	}, {
		"obj1": &Value{s: "20", t: TypeNumber},
		"obj2": &Value{s: "100", t: TypeNumber},
		"op":   ">",
		"exp":  false,
		"err":  nil,
	},
}

func Test_jsonpath_cmp_any(t *testing.T) {
	for idx, tcase := range tcase_cmp_any {
		//for idx, tcase := range tcase_cmp_any[8:] {
		t.Logf("idx: %v, %v %v %v, exp: %v", idx, tcase["obj1"], tcase["op"], tcase["obj2"], tcase["exp"])
		res, err := cmp_any(tcase["obj1"].(*Value), tcase["obj2"].(*Value), tcase["op"].(string))
		exp := tcase["exp"].(bool)
		exp_err := tcase["err"]
		if exp_err != nil {
			if err == nil {
				t.Errorf("idx: %d error not raised: %v(exp)", idx, exp_err)
				break
			}
		} else {
			if err != nil {
				t.Errorf("idx: %v, error: %v", idx, err)
				break
			}
		}
		if res != exp {
			t.Errorf("idx: %v, %v(got) != %v(exp)", idx, res, exp)
			break
		}
	}
}

func Test_jsonpath_string_equal(t *testing.T) {
	data := `{
   "store": {
       "book": [
           {
               "category": "reference",
               "author": "Nigel Rees",
               "title": "Sayings of the Century",
               "price": 8.95
           },
           {
               "category": "fiction",
               "author": "Evelyn Waugh",
               "title": "Sword of Honour",
               "price": 12.99
           },
           {
               "category": "fiction",
               "author": "Herman Melville",
               "title": "Moby Dick",
               "isbn": "0-553-21311-3",
               "price": 8.99
           },
           {
               "category": "fiction",
               "author": "J. R. R. Tolkien",
               "title": "The Lord of the Rings",
               "isbn": "0-395-19395-8",
               "price": 22.99
           }
       ],
       "bicycle": {
           "color": "red",
           "price": 19.95
       }
   },
   "expensive": 10
}`

	var j *Value
	j = MustParse(data)

	res, err := JsonPathLookup(j, "$.store.book[?(@.author == 'Nigel Rees')].price")
	t.Log(res, err)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if fmt.Sprintf("%v", res) != "[8.95]" {
		t.Fatalf("not the same: %v", res)
	}
}

func Test_jsonpath_null_in_the_middle(t *testing.T) {
	data := `{"head_commit": null}`

	var j *Value
	j = MustParse(data)

	res, err := JsonPathLookup(j, "$.head_commit.author.username")
	t.Log(res, err)
}

func Test_jsonpath_num_cmp(t *testing.T) {
	data := `{
	"books": [
        { "name": "My First Book", "price": 10 },
		{ "name": "My Second Book", "price": 20 }
		]
}`
	var j *Value
	j = MustParse(data)

	res, err := JsonPathLookup(j, "$.books[?(@.price > 100)].name")
	if err != nil {
		t.Fatal(err)
	}

	arr, _ := res.ToArray()
	if len(arr) != 0 {
		t.Fatal("should return [], got: ", arr)
	}

}

func BenchmarkJsonPathLookupCompiled(b *testing.B) {
	c, err := Compile("$.store.book[0].price")
	if err != nil {
		b.Fatalf("%v", err)
	}
	for n := 0; n < b.N; n++ {
		res, err := c.Lookup(json_data)
		if res_v, ok := res.ToFloat64(); ok != nil || res_v != 8.95 {
			b.Errorf("$.store.book[0].price should be 8.95")
		}
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJsonPathLookup(b *testing.B) {
	for n := 0; n < b.N; n++ {
		res, err := JsonPathLookup(json_data, "$.store.book[0].price")
		if res_v, ok := res.ToFloat64(); ok != nil || res_v != 8.95 {
			b.Errorf("$.store.book[0].price should be 8.95")
		}
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkJsonPathLookup_0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.expensive")
	}
}

func BenchmarkJsonPathLookup_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[0].price")
	}
}

func BenchmarkJsonPathLookup_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[-1].price")
	}
}

func BenchmarkJsonPathLookup_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[0,1].price")
	}
}

func BenchmarkJsonPathLookup_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[0:2].price")
	}
}

func BenchmarkJsonPathLookup_5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.isbn)].price")
	}
}

func BenchmarkJsonPathLookup_6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.price > 10)].title")
	}
}

func BenchmarkJsonPathLookup_7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.price < $.expensive)].price")
	}
}

func BenchmarkJsonPathLookup_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[:].price")
	}
}

func BenchmarkJsonPathLookup_9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.author == 'Nigel Rees')].price")
	}
}

func BenchmarkJsonPathLookup_10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		JsonPathLookup(json_data, "$.store.book[?(@.author =~ /(?i).*REES/)].price")
	}
}

func TestReg(t *testing.T) {
	r := regexp.MustCompile(`(?U).*REES`)
	t.Log(r)
	t.Log(r.Match([]byte(`Nigel Rees`)))

	res, err := JsonPathLookup(json_data, "$.store.book[?(@.author =~ /(?i).*REES/ )].author")
	t.Log(err, res)

	author := res.MustArray()[0].MustString()
	t.Log(author)
	if author != "Nigel Rees" {
		t.Fatal("should be `Nigel Rees` but got: ", author)
	}
}

var tcases_reg_op = []struct {
	Line string
	Exp  string
	Err  bool
}{
	{``, ``, true},
	{`xxx`, ``, true},
	{`/xxx`, ``, true},
	{`xxx/`, ``, true},
	{`'/xxx/'`, ``, true},
	{`"/xxx/"`, ``, true},
	{`/xxx/`, `xxx`, false},
	{`/π/`, `π`, false},
}

func TestRegOp(t *testing.T) {
	for idx, tcase := range tcases_reg_op {
		fmt.Println("idx: ", idx, "tcase: ", tcase)
		res, err := regFilterCompile(tcase.Line)
		if tcase.Err == true {
			if err == nil {
				t.Fatal("expect err but got nil")
			}
		} else {
			if res == nil || res.String() != tcase.Exp {
				t.Fatal("different. res:", res)
			}
		}
	}
}

func Test_jsonpath_rootnode_is_array(t *testing.T) {
	data := `[{
   "test": 12.34
}, {
	"test": 13.34
}, {
	"test": 14.34
}]
`

	j := MustParse(data)

	res, err := JsonPathLookup(j, "$[0].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil || res.MustFloat64() != 12.34 {
		t.Fatalf("different:  res:%v, exp: 123", res)
	}
}

func Test_jsonpath_rootnode_is_array_range(t *testing.T) {
	data := `[{
   "test": 12.34
}, {
	"test": 13.34
}, {
	"test": 14.34
}]
`

	j := MustParse(data)

	res, err := JsonPathLookup(j, "$[:1].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil {
		t.Fatal("res is nil")
	}
	ares := res.MustArray()
	for idx, v := range ares {
		t.Logf("idx: %v, v: %v", idx, v)
	}
	if len(ares) != 2 {
		t.Fatalf("len is not 2. got: %v", len(ares))
	}
	if ares[0].MustFloat64() != 12.34 {
		t.Fatalf("idx: 0, should be 12.34. got: %v", ares[0])
	}
	if ares[1].MustFloat64() != 13.34 {
		t.Fatalf("idx: 0, should be 12.34. got: %v", ares[1])
	}
}

func Test_jsonpath_rootnode_is_nested_array(t *testing.T) {
	data := `[ [ {"test":1.1}, {"test":2.1} ], [ {"test":3.1}, {"test":4.1} ] ]`

	j := MustParse(data)

	res, err := JsonPathLookup(j, "$[0].[0].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil || res.MustFloat64() != 1.1 {
		t.Fatalf("different:  res:%v, exp: 123", res)
	}
}

func Test_jsonpath_rootnode_is_nested_array_range(t *testing.T) {
	data := `[ [ {"test":1.1}, {"test":2.1} ], [ {"test":3.1}, {"test":4.1} ] ]`

	j := MustParse(data)

	res, err := JsonPathLookup(j, "$[:1].[0].test")
	t.Log(res, err)
	if err != nil {
		t.Fatal("err:", err)
	}
	if res == nil {
		t.Fatal("res is nil")
	}
	ares := res.MustArray()
	for idx, v := range ares {
		t.Logf("idx: %v, v: %v", idx, v)
	}
	if len(ares) != 2 {
		t.Fatalf("len is not 2. got: %v", len(ares))
	}

	//FIXME: `$[:1].[0].test` got wrong result
	if ares[0].MustFloat64() != 1.1 {
		t.Fatalf("idx: 0, should be 1.1, got: %v", ares[0])
	}
	if ares[1].MustFloat64() != 3.1 {
		t.Fatalf("idx: 0, should be 3.1, got: %v", ares[1])
	}
}

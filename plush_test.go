package plush

import (
	"errors"
	"fmt"
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Render_Simple_HTML(t *testing.T) {
	r := require.New(t)

	input := `<p>Hi</p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal(input, s)
}

func Test_Render_HTML_InjectedString(t *testing.T) {
	r := require.New(t)

	input := `<p><%= "mark" %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>mark</p>", s)
}

func Test_Render_EscapedString(t *testing.T) {
	r := require.New(t)

	input := `<p><%= "<script>alert('pwned')</script>" %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>&lt;script&gt;alert(&#39;pwned&#39;)&lt;/script&gt;</p>", s)
}

func Test_Render_Injected_Variable(t *testing.T) {
	r := require.New(t)

	input := `<p><%= name %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"name": "Mark",
	}))
	r.NoError(err)
	r.Equal("<p>Mark</p>", s)
}

func Test_Render_Let_Hash(t *testing.T) {
	r := require.New(t)

	input := `<p><% let h = {"a": "A"} %><%= h["a"] %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>A</p>", s)
}

func Test_Render_Hash_Array_Index(t *testing.T) {
	r := require.New(t)

	input := `<%= m["first"] + " " + m["last"] %>|<%= a[0+1] %>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"m": map[string]string{"first": "Mark", "last": "Bates"},
		"a": []string{"john", "paul"},
	}))
	r.NoError(err)
	r.Equal("Mark Bates|paul", s)
}

func Test_Render_Missing_Variable(t *testing.T) {
	r := require.New(t)

	input := `<p><%= name %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p></p>", s)
}

func Test_Render_Function_Call(t *testing.T) {
	r := require.New(t)

	input := `<p><%= f() %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"f": func() string {
			return "hi!"
		},
	}))
	r.NoError(err)
	r.Equal("<p>hi!</p>", s)
}

func Test_Render_Function_Call_With_Arg(t *testing.T) {
	r := require.New(t)

	input := `<p><%= f("mark") %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"f": func(s string) string {
			return fmt.Sprintf("hi %s!", s)
		},
	}))
	r.NoError(err)
	r.Equal("<p>hi mark!</p>", s)
}

func Test_Render_Function_Call_With_Variable_Arg(t *testing.T) {
	r := require.New(t)

	input := `<p><%= f(name) %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"f": func(s string) string {
			return fmt.Sprintf("hi %s!", s)
		},
		"name": "mark",
	}))
	r.NoError(err)
	r.Equal("<p>hi mark!</p>", s)
}

func Test_Render_Function_Call_With_Hash(t *testing.T) {
	r := require.New(t)

	input := `<p><%= f({name: name}) %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"f": func(m map[string]interface{}) string {
			return fmt.Sprintf("hi %s!", m["name"])
		},
		"name": "mark",
	}))
	r.NoError(err)
	r.Equal("<p>hi mark!</p>", s)
}

func Test_Render_HTML_Escape(t *testing.T) {
	r := require.New(t)

	input := `<%= safe() %>|<%= unsafe() %>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"safe": func() string {
			return "<b>unsafe</b>"
		},
		"unsafe": func() template.HTML {
			return "<b>unsafe</b>"
		},
	}))
	r.NoError(err)
	r.Equal("&lt;b&gt;unsafe&lt;/b&gt;|<b>unsafe</b>", s)
}

func Test_Render_Function_Call_With_Error(t *testing.T) {
	r := require.New(t)

	input := `<p><%= f() %></p>`
	_, err := Render(input, NewContextWith(map[string]interface{}{
		"f": func() (string, error) {
			return "hi!", errors.New("oops!")
		},
	}))
	r.Error(err)
}

func Test_Render_Function_Call_With_Block(t *testing.T) {
	r := require.New(t)

	input := `<p><%= f() { %>hello<% } %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"f": func(h HelperContext) string {
			s, _ := h.Block()
			return s
		},
	}))
	r.NoError(err)
	r.Equal("<p>hello</p>", s)
}

type greeter struct{}

func (g greeter) Greet(s string) string {
	return fmt.Sprintf("hi %s!", s)
}

func Test_Render_Function_Call_On_Callee(t *testing.T) {
	r := require.New(t)

	input := `<p><%= g.Greet("mark") %></p>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"g": greeter{},
	}))
	r.NoError(err)
	r.Equal(`<p>hi mark!</p>`, s)
}

func Test_Render_For_Array(t *testing.T) {
	r := require.New(t)
	input := `<% for (i,v) in ["a", "b", "c"] {return v} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("", s)
}

func Test_Render_For_Hash(t *testing.T) {
	r := require.New(t)
	input := `<%= for (k,v) in myMap { %><%= k + ":" + v%><% } %>`
	s, err := Render(input, NewContextWith(map[string]interface{}{
		"myMap": map[string]string{
			"a": "A",
			"b": "B",
		},
	}))
	r.NoError(err)
	r.Contains(s, "a:A")
	r.Contains(s, "b:B")
}

func Test_Render_For_Array_Return(t *testing.T) {
	r := require.New(t)
	input := `<%= for (i,v) in ["a", "b", "c"] {return v} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("abc", s)
}

func Test_Render_For_Array_Key_Only(t *testing.T) {
	r := require.New(t)
	input := `<%= for (v) in ["a", "b", "c"] {%><%=v%><%} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("abc", s)
}

func Test_Render_For_Func_Range(t *testing.T) {
	r := require.New(t)
	input := `<%= for (v) in range(3,5) { %><%=v%><% } %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("345", s)
}

func Test_Render_For_Func_Between(t *testing.T) {
	r := require.New(t)
	input := `<%= for (v) in between(3,6) { %><%=v%><% } %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("45", s)
}

func Test_Render_For_Func_Until(t *testing.T) {
	r := require.New(t)
	input := `<%= for (v) in until(3) { %><%=v%><% } %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("012", s)
}

func Test_Render_For_Array_Key_Value(t *testing.T) {
	r := require.New(t)
	input := `<%= for (i,v) in ["a", "b", "c"] {%><%=i%><%=v%><%} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("0a1b2c", s)
}

func Test_Render_If(t *testing.T) {
	r := require.New(t)
	input := `<% if (true) { return "hi"} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("", s)
}

func Test_Render_If_Return(t *testing.T) {
	r := require.New(t)
	input := `<%= if (true) { return "hi"} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("hi", s)
}

func Test_Render_If_Return_HTML(t *testing.T) {
	r := require.New(t)
	input := `<%= if (true) { %>hi<%} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("hi", s)
}

func Test_Render_If_And(t *testing.T) {
	r := require.New(t)
	input := `<%= if (false && true) { %> hi <%} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("", s)
}

func Test_Render_If_Or(t *testing.T) {
	r := require.New(t)
	input := `<%= if (false || true) { %>hi<%} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("hi", s)
}

func Test_Render_If_Nil(t *testing.T) {
	r := require.New(t)
	input := `<%= if (names && len(names) >= 1) { %>hi<%} %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("", s)
}

func Test_Render_If_Else_Return(t *testing.T) {
	r := require.New(t)
	input := `<p><%= if (false) { return "hi"} else { return "bye"} %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>bye</p>", s)
}

func Test_Render_If_LessThan(t *testing.T) {
	r := require.New(t)
	input := `<p><%= if (1 < 2) { return "hi"} else { return "bye"} %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>hi</p>", s)
}

func Test_Render_If_BangFalse(t *testing.T) {
	r := require.New(t)
	input := `<p><%= if (!false) { return "hi"} else { return "bye"} %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>hi</p>", s)
}

func Test_Render_If_NotEq(t *testing.T) {
	r := require.New(t)
	input := `<p><%= if (1 != 2) { return "hi"} else { return "bye"} %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>hi</p>", s)
}

func Test_Render_If_GtEq(t *testing.T) {
	r := require.New(t)
	input := `<p><%= if (1 >= 2) { return "hi"} else { return "bye"} %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>bye</p>", s)
}

func Test_Render_If_Else_True(t *testing.T) {
	r := require.New(t)
	input := `<p><%= if (true) { %>hi<% } else { %>bye<% } %></p>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("<p>hi</p>", s)
}

func Test_Render_ShowNoShow(t *testing.T) {
	r := require.New(t)
	input := `<%= "shown" %><% "notshown" %>`
	s, err := Render(input, NewContext())
	r.NoError(err)
	r.Equal("shown", s)
}

func Test_Render_Struct_Attribute(t *testing.T) {
	r := require.New(t)
	input := `<%= f.Name %>`
	ctx := NewContext()
	f := struct {
		Name string
	}{"Mark"}
	ctx.Set("f", f)
	s, err := Render(input, ctx)
	r.NoError(err)
	r.Equal("Mark", s)
}

func Test_Render_ScriptFunction(t *testing.T) {
	r := require.New(t)

	input := `<% let add = fn(x) { return x + 2; }; %><%= add(2) %>`

	s, err := Render(input, NewContext())
	if err != nil {
		r.NoError(err)
	}
	r.Equal("4", s)
}
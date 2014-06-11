package jantar

import (
	"fmt"
	"github.com/tsurai/jantar/context"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type route struct {
	cName   string
	cAction string
	pattern string
	method  string
	handler http.HandlerFunc
}

type rootNode struct {
	get    *pathNode
	post   *pathNode
	delete *pathNode
	put    *pathNode
}

type pathLeaf struct {
	paramNames []string
	route      *route
}

type pathNode struct {
	edges    map[string]*pathNode
	wildcard *pathNode
	leaf     *pathLeaf
}

type router struct {
	namedRoutes map[string]*route
	pathRoot    rootNode
}

// Router functions ----------------------------------------------
func newRouter() *router {
	return &router{namedRoutes: make(map[string]*route), pathRoot: rootNode{newPathNode(), newPathNode(), newPathNode(), newPathNode()}}
}

func (r *router) getMethodPathNode(method string) *pathNode {
	switch strings.ToUpper(method) {
	case "GET":
		return r.pathRoot.get
	case "POST":
		return r.pathRoot.post
	case "PUT":
		return r.pathRoot.put
	case "DELETE":
		return r.pathRoot.delete
	default:
		return r.pathRoot.get
	}
}

func (r *router) findPathLeaf(method string, path string) (*pathLeaf, map[string]string) {
	var variables []string
	node := r.getMethodPathNode(method)

	for _, segment := range splitPath(path) {
		if edge, ok := node.edges[segment]; ok {
			node = edge
		} else {
			if node.wildcard == nil {
				return nil, nil
			}
			variables = append(variables, segment)
			node = node.wildcard
		}
	}

	if node.leaf == nil {
		return nil, nil
	}

	param := make(map[string]string)
	for i, v := range variables {
		param[node.leaf.paramNames[i]] = v
	}

	return node.leaf, param
}

func (r *router) insertPathLeaf(method string, path string) *pathLeaf {
	var paramNames []string
	node := r.getMethodPathNode(method)

	for _, segment := range splitPath(path) {
		if strings.HasPrefix(segment, ":") {
			if node.wildcard == nil {
				node.wildcard = newPathNode()
			}

			paramNames = append(paramNames, segment[1:])
			node = node.wildcard
			continue
		}

		if edge, ok := node.edges[segment]; ok {
			node = edge
		} else {
			node.edges[segment] = newPathNode()
			node = node.edges[segment]
		}
	}

	node.leaf = &pathLeaf{paramNames, nil}
	return node.leaf
}

func (r *router) addRoute(method string, path string, handler interface{}) *route {
	route := newRoute(strings.ToUpper(method), path, handler)

	node := r.insertPathLeaf(method, path)
	node.route = route

	// is route a controller route
	if route.cName != "" {
		// add to named routes with name as controller#action
		r.namedRoutes[strings.ToLower(route.cName+"#"+route.cAction)] = route
	}

	return route
}

func (r *router) searchRoute(req *http.Request) *route {
	if node, params := r.findPathLeaf(req.Method, req.URL.Path); node != nil {
		if len(params) != 0 {
			context.Set(req, "_UrlParam", params, true)
		}

		return node.route
	}

	return nil
}

func (r *router) getReverseURL(name string, param []interface{}) string {
	route := r.getNamedRoute(name)
	nParam := len(param)

	if route != nil {
		i := -1
		regex := regexp.MustCompile(":[^/]+")
		url := regex.ReplaceAllStringFunc(route.pattern, func(str string) string {
			i = i + 1
			if i < nParam {
				return fmt.Sprintf("%v", param[i])
			}
			return ""
		})

		return url
	}

	return ""
}

func (r *router) getNamedRoute(name string) *route {
	if route, ok := r.namedRoutes[strings.ToLower(name)]; ok {
		return route
	}

	return nil
}

// Route functions ---------------------------------------------
func newRoute(method string, pattern string, handler interface{}) *route {
	var finalFunc http.HandlerFunc
	cName := ""
	cAction := ""

	if reflect.TypeOf(handler) == reflect.TypeOf(http.NotFound) {
		finalFunc = handler.(func(http.ResponseWriter, *http.Request))
	} else if cType := getControllerType(handler); cType != nil {
		fn := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
		if fn == nil {
			Log.Warning("failed to add route. Can't fetch controller function")
			return nil
		}

		regex := regexp.MustCompile(".*\\.\\(\\*(.*)\\)\\.(.*)")
		matches := regex.FindStringSubmatch(fn.Name())

		if len(matches) == 3 && matches[0] == fn.Name() {
			cName = matches[1]
			cAction = matches[2]
		}

		finalFunc = func(rw http.ResponseWriter, r *http.Request) {
			c := newController(cType, rw, r, cName, cAction)

			var in []reflect.Value
			in = append(in, reflect.ValueOf(c))

			reflect.ValueOf(handler).Call(in)
		}
	} else {
		Log.Warningd(JLData{"type": reflect.TypeOf(handler), "wanted": reflect.TypeOf(http.NotFound)}, "failed to add route. Invalid handler type")
	}

	return &route{cName, cAction, pattern, method, finalFunc}
}

func (r *route) Name(name string) {
	router := context.GetGlobal("Router").(*router)
	router.namedRoutes[strings.ToLower(name)] = r
}

// Helper functions ---------------------------------------------
func newPathNode() *pathNode {
	return &pathNode{edges: make(map[string]*pathNode)}
}

func splitPath(path string) []string {
	segments := strings.Split(path, "/")

	if segments[0] == "" {
		segments = segments[1:]
	}
	return segments
}

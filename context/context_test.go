package context

import (
	"net/http"
	"testing"
)

var req *http.Request

func TestAdd(t *testing.T) {
	assertEqual := func(val interface{}, exp interface{}) {
		if val != exp {
			t.Errorf("Expected %v, got %v.", exp, val)
		}
	}

	/* request data set */
	{
		Set(req, "key", "item", false)
		value, ok := GetOk(req, "key")
		assertEqual(value, "item")
		assertEqual(ok, true)
	}

	/* global data set */
	{
		SetGlobal("key", "item", false)
		value, ok := GetOk(req, "key")
		assertEqual(value, "item")
		assertEqual(ok, true)
	}

	/* test correct map size */
	Set(req, "foo", "bar", false)
	SetGlobal("foo", "bar", false)
	assertEqual(len(globalData), 2)
	assertEqual(len(requestData[req]), 2)

	/* test invalid access on request data */
	{
		var ok bool
		value := Get(req, "invalid")
		assertEqual(value, nil)

		value, ok = GetOk(req, "invalid")
		assertEqual(value, nil)
		assertEqual(ok, false)
	}

	/* test invalid access on global data */
	{
		var ok bool
		value := GetGlobal("invalid")
		assertEqual(value, nil)

		value, ok = GetGlobalOk("invalid")
		assertEqual(value, nil)
		assertEqual(ok, false)
	}
}

func TestAddReadOnly(t *testing.T) {
	assertEqual := func(val interface{}, exp interface{}) {
		if val != exp {
			t.Errorf("Expected %v, got %v.", exp, val)
		}
	}

	/* test request data set */
	Set(req, "baz", "bar", true)
	Set(req, "baz", "foo", true)
	assertEqual(Get(req, "baz"), "bar")

	/* test global data set */
	SetGlobal("baz", "bar", true)
	SetGlobal("baz", "foo", true)
	assertEqual(GetGlobal("baz"), "bar")
}

func TestClearData(t *testing.T) {
	ClearData(req)
	if _, ok := requestData[req]; ok {
		t.Errorf("Failed to clear request data")
	}
}

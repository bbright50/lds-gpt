package utils

import "fmt"

func JSONAccessor(value any) JSON {
	return jsonWrapper{v: value, err: nil}
}

type JSON interface {
	Get(key string) JSON
	IndexArray(i int) JSON
	Value() (any, error)
	ValueAsArray() ([]any, error)
	ValueAsMap() (map[string]any, error)
}

type jsonWrapper struct {
	v   any
	err error
}

func (j jsonWrapper) Get(key string) JSON {
	valueMap, ok := j.v.(map[string]any)
	if !ok {
		return jsonWrapper{
			err: fmt.Errorf("cannot get value for key %s: value is not a map", key),
		}
	}
	value, ok := valueMap[key]
	if !ok {
		return jsonWrapper{
			err: fmt.Errorf("cannot get value for key %s: key not found", key),
		}
	}
	return jsonWrapper{v: value, err: nil}
}

func (j jsonWrapper) IndexArray(i int) JSON {
	valueArray, ok := j.v.([]any)
	if !ok {
		return jsonWrapper{
			err: fmt.Errorf("cannot get value for index %d: value is not an array", i),
		}
	}
	if i < 0 || i >= len(valueArray) {
		return jsonWrapper{
			err: fmt.Errorf("cannot get value for index %d: index out of bounds", i),
		}
	}
	return jsonWrapper{v: valueArray[i], err: nil}
}

func (j jsonWrapper) Value() (any, error) {
	if j.err != nil {
		return nil, j.err
	}
	return j.v, nil
}

func (j jsonWrapper) ValueAsMap() (map[string]any, error) {
	if j.err != nil {
		return nil, j.err
	}
	valueMap, ok := j.v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("value is not a map")
	}
	return valueMap, nil
}

func (j jsonWrapper) ValueAsArray() ([]any, error) {
	if j.err != nil {
		return nil, j.err
	}
	valueArray, ok := j.v.([]any)
	if !ok {
		return nil, fmt.Errorf("value is not an array")
	}
	return valueArray, nil
}

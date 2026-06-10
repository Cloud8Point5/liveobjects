package liveobjects

import (
	"errors"
	"reflect"
	"strings"
)

type TransitionStepType int

const (
	ReplaceStep TransitionStepType = 0
	DeleteStep  TransitionStepType = 1
)

type TranstionStep struct {
	Type  TransitionStepType `codec:"type"`
	Path  string             `codec:"path"`
	Value any        `codec:"value,omitempty"`
}

func GetObjectTransition(oldValue any, newValue any) []TranstionStep {
	oldType := reflect.TypeOf(oldValue)
	newType := reflect.TypeOf(newValue)
	if oldValue == nil {
		if newValue == nil {
			return []TranstionStep{}
		} else {
			return []TranstionStep{
				{
					Type:  ReplaceStep,
					Path:  "",
					Value: newValue,
				},
			}
		}
	}
	if newValue == nil || oldType.Kind() != newType.Kind() {
		return []TranstionStep{
			{
				Type:  ReplaceStep,
				Path:  "",
				Value: newValue,
			},
		}
	}
	if oldType.Comparable() && newType.Comparable() {
		if oldValue == newValue {
			return []TranstionStep{}
		}
	}
	switch oldType.Kind() {
	case reflect.Map:
		stepList := make([]TranstionStep, 0)

		oldMap := oldValue.(map[string]any)
		newMap := newValue.(map[string]any)

		for k, v := range oldMap {
			if newV, exists := newMap[k]; exists {
				diff := GetObjectTransition(v, newV)
				for i, step := range diff {
					if step.Path == "" {
						step.Path = k
					} else {
						step.Path = k + "/" + step.Path
					}
					diff[i] = step
				}

				stepList = append(stepList, diff...)
			} else {
				stepList = append(stepList, TranstionStep{
					Type: DeleteStep,
					Path: k,
				})
			}
		}

		for k := range newMap {
			if _, exists := oldMap[k]; !exists {
				stepList = append(stepList, TranstionStep{
					Type:  ReplaceStep,
					Path:  k,
					Value: newMap[k],
				})
			}
		}

		return stepList
	case reflect.Slice:
		oldSlice := oldValue.([]any)
		newSlice := newValue.([]any)
		if len(oldSlice) != len(newSlice) {
			return []TranstionStep{
				{
					Type:  ReplaceStep,
					Path:  "",
					Value: newValue,
				},
			}
		} else {
			for i := 0; i < len(oldSlice); i++ {
				transition := GetObjectTransition(oldSlice[i], newSlice[i])
				if len(transition) != 0 {
					return []TranstionStep{
						{
							Type:  ReplaceStep,
							Path:  "",
							Value: newValue,
						},
					}
				} else {
					return []TranstionStep{}
				}
			}
		}
	default:
		return []TranstionStep{
			{
				Type:  ReplaceStep,
				Path:  "",
				Value: newValue,
			},
		}
	}

	// should never get here because of the default block above
	return []TranstionStep{}
}

func ApplyObjectTransition(oldValue any, steps []TranstionStep) (any, error) {

	if len(steps) == 0 {
		return oldValue, nil
	}
	for _, step := range steps {
		switch step.Type {
		case ReplaceStep:
			if step.Path == "" {
				return step.Value, nil
			}
			oldMap, ok := oldValue.(map[string]any)
			if !ok {
				// it might be something decoded from messagepack, which would be map[any]any
				oldMapAny, ok := oldValue.(map[any]any)
				if !ok {
					return nil, errors.New("replace step includes non-empty path, yet old value is not a map")
				}
				oldMap = StringKeys(oldMapAny)
				oldValue = oldMap
			}
			segments := strings.Split(step.Path, "/")
			lastSegment := len(segments) - 1
			for i, segment := range segments {
				if i == lastSegment {
					oldMap[segment] = step.Value
				} else {
					tempOldMap, ok := oldMap[segment].(map[string]any)
					if !ok {
						// it might be something decoded from messagepack, which would be map[any]any
						oldMapAny, ok := oldMap[segment].(map[any]any)
						if !ok {
							return nil, errors.New("replace step segment leads to a non-map")
						}
						tempOldMap = StringKeys(oldMapAny)
						oldMap[segment] = tempOldMap
					}
					oldMap = tempOldMap
				}
			}
		case DeleteStep:
			if step.Path == "" {
				return nil, errors.New("delete step path cannot be empty")
			}
			oldMap, ok := oldValue.(map[string]any)
			if !ok {
				// it might be something decoded from messagepack, which would be map[any]any
				oldMapAny, ok := oldValue.(map[any]any)
				if !ok {
					return nil, errors.New("delete step old value is not a map")
				}
				oldMap = StringKeys(oldMapAny)
				oldValue = oldMap
			}
			segments := strings.Split(step.Path, "/")
			lastSegment := len(segments) - 1
			for i, segment := range segments {
				if i == lastSegment {
					delete(oldMap, segment)
				} else {
					tempOldMap, ok := oldMap[segment].(map[string]any)
					if !ok {
						// it might be something decoded from messagepack, which would be map[any]any
						oldMapAny, ok := oldMap[segment].(map[any]any)
						if !ok {
							return nil, errors.New("delete step segment leads to a non-map")
						}
						tempOldMap = StringKeys(oldMapAny)
						oldMap[segment] = tempOldMap
					}
					oldMap = tempOldMap
				}
			}
		}
	}
	return oldValue, nil
}

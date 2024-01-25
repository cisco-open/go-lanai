package log

import "runtime"

type Stacktracer func() (frames []*runtime.Frame, fallback interface{})

// RuntimeStacktracer find stacktrace frames with runtime package
// skip: skip certain number of stacks from the top, including the call to the Stacktracer itself
// depth: max number of stack frames to extract
func RuntimeStacktracer(skip int, depth int) Stacktracer {
    return func() (frames []*runtime.Frame, fallback interface{}) {
        rpc := make([]uintptr, depth)
        count := runtime.Callers(skip, rpc)
        rawFrames := runtime.CallersFrames(rpc)
        frames = make([]*runtime.Frame, count)
        for i := 0; i < count; i++ {
            frame, more := rawFrames.Next()
            frames[i] = &frame
            if !more {
                break
            }
        }
        return frames, nil
    }
}

// RuntimeCaller equivalent to RuntimeStacktracer(skip, 1)
func RuntimeCaller(skip int) Stacktracer {
    return RuntimeStacktracer(skip, 1)
}


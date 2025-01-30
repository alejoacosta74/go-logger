package logger

import (
	"path/filepath"
	"runtime"
	"strings"
)

// callerInfo holds the extracted runtime caller information
type callerInfo struct {
	funcName  string
	fileName  string
	line      int
	pkgName   string
	shortFunc string
}

// extractCallerInfo without anonymous function filtering
func extractCallerInfo(skipFrames int) (callerInfo, bool) {
	var info callerInfo
	for i := skipFrames; i < skipFrames+15; i++ {
		if pc, file, line, ok := runtime.Caller(i); ok {
			funcName := runtime.FuncForPC(pc).Name()
			if !strings.Contains(funcName, "logrus") &&
				!strings.Contains(funcName, "runtime.") &&
				!strings.Contains(funcName, "testing.") &&
				!strings.Contains(file, "runtime/") &&
				!strings.Contains(file, "testing/") &&
				!strings.Contains(file, "logger.go") &&
				!strings.Contains(funcName, "WithRuntimeContext") {

				info.funcName = funcName

				var fileName string
				fileParts := strings.Split(file, string(filepath.Separator))
				if len(fileParts) >= 2 {
					fileName = filepath.Join(fileParts[len(fileParts)-2], fileParts[len(fileParts)-1])
				} else {
					fileName = fileParts[len(fileParts)-1]
				}
				info.fileName = fileName
				info.line = line

				lastDot := strings.LastIndex(funcName, ".")
				if lastDot != -1 {
					pkgPath := funcName[:lastDot]
					fullFunc := funcName[lastDot+1:]
					pkgParts := strings.Split(pkgPath, "/")
					info.pkgName = pkgParts[len(pkgParts)-1]
					info.shortFunc = fullFunc
					return info, true
				}
			}
		}
	}
	return info, false
}

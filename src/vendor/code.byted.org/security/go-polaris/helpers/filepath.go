package helpers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Benchmark 压测结果
// goos: darwin
// goarch: amd64
// pkg: code.byted.org/security/go-polaris/helpers
//
// BenchmarkGetRealFilePath
// BenchmarkGetRealFilePath           	 5504766	       203 ns/op
// BenchmarkGetRealFilePath-2         	 6864664	       173 ns/op
// BenchmarkGetRealFilePath-4         	 6666596	       173 ns/op
// BenchmarkGetRealFilePath-8         	 7015263	       173 ns/op
// BenchmarkGetRealFilePath-12        	 5815892	       174 ns/op
//
// BenchmarkGetRealFileName
// BenchmarkGetRealFileName       	259870861	         4.28 ns/op
// BenchmarkGetRealFileName-2     	285352750	         4.28 ns/op
// BenchmarkGetRealFileName-4     	272578628	         4.47 ns/op
// BenchmarkGetRealFileName-8     	262621897	         4.40 ns/op
// BenchmarkGetRealFileName-12    	269856204	         4.45 ns/op

var (
	//ErrFilepathNotAllowed 文件路径不符合要求
	ErrFilepathNotAllowed = errors.New("file path is not allowed")
)

//GetRealFilePath 判断绝对路径是否符合预期
// 参数allowAbsDirs中的路径最好为绝对路径
func GetRealFilePath(filePath string, allowAbsDirs ...string) (realPath string, err error) {
	if filePath == "" {
		return filePath, nil
	}

	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) {
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			return "", errors.Wrap(err, "parse the absolute file path")
		}
	}

	dir, file := filepath.Split(filePath)
	for _, allowDir := range allowAbsDirs {
		if !filepath.IsAbs(allowDir) {
			allowDir, err = filepath.Abs(allowDir)
			if err != nil {
				return "", errors.Wrap(err, fmt.Sprintf("parse allow dir %q fail", allowDir))
			}
		}

		if strings.HasPrefix(dir, filepath.Clean(allowDir)) {
			return filepath.Join(dir, file), nil
		}
	}
	return "", ErrFilepathNotAllowed
}

//GetRealFileName 获取filename
func GetRealFileName(filePath string) (realName string) {
	_, realName = filepath.Split(filePath)
	return realName
}

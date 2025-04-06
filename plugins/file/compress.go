package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// compressDirectory 压缩目录
func (p *FileTransferPlugin) compressDirectory(src, dest string) error {
	// 检查目标文件扩展名
	ext := strings.ToLower(filepath.Ext(dest))

	switch ext {
	case ".zip":
		return p.zipDirectory(src, dest)
	case ".gz", ".tgz":
		return p.tarGzDirectory(src, dest)
	default:
		return p.zipDirectory(src, dest)
	}
}

// zipDirectory 使用zip压缩目录
func (p *FileTransferPlugin) zipDirectory(src, dest string) error {
	// 创建zip文件
	zipFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历源目录
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建zip头信息
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 设置相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		header.Name = relPath

		// 设置压缩方法
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		// 创建writer
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是目录，跳过
		if info.IsDir() {
			return nil
		}

		// 打开源文件
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 复制文件内容
		_, err = io.Copy(writer, file)
		return err
	})
}

// tarGzDirectory 使用tar.gz压缩目录
func (p *FileTransferPlugin) tarGzDirectory(src, dest string) error {
	// 创建目标文件
	tarFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	// 创建gzip writer
	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	// 创建tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 遍历源目录
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建tar头信息
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// 设置相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		header.Name = relPath

		// 写入头信息
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是目录，跳过
		if info.IsDir() {
			return nil
		}

		// 打开源文件
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 复制文件内容
		_, err = io.Copy(tarWriter, file)
		return err
	})
}

// decompressFile 解压文件
func (p *FileTransferPlugin) decompressFile(src, dest string) error {
	// 检查源文件扩展名
	ext := strings.ToLower(filepath.Ext(src))

	switch ext {
	case ".zip":
		return p.unzipFile(src, dest)
	case ".gz", ".tgz":
		return p.untarGzFile(src, dest)
	default:
		return fmt.Errorf("unsupported archive format: %s", ext)
	}
}

// unzipFile 解压zip文件
func (p *FileTransferPlugin) unzipFile(src, dest string) error {
	// 打开zip文件
	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// 遍历zip文件中的所有文件
	for _, file := range zipReader.File {
		// 构建目标路径
		path := filepath.Join(dest, file.Name)

		// 检查路径是否在目标目录内
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		// 如果是目录，创建目录
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 创建目标文件
		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		// 打开源文件
		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		// 复制文件内容
		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// untarGzFile 解压tar.gz文件
func (p *FileTransferPlugin) untarGzFile(src, dest string) error {
	// 打开源文件
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// 创建tar reader
	tarReader := tar.NewReader(gzipReader)

	// 遍历tar文件中的所有文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 构建目标路径
		path := filepath.Join(dest, header.Name)

		// 检查路径是否在目标目录内
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		// 根据文件类型处理
		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			// 创建文件
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// 复制文件内容
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}
	}

	return nil
}

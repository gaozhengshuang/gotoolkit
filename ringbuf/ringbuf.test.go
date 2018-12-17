/// @file ringbuf_test.go
/// @brief 
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-12-11

//package main
//import "fmt"
package ringbuf
import "fmt"

func RunRingBufTest() {
	testRingBufGrow()
	testRingBufWrite()
	testRingBufRead()
}

func printfBuf(buf *Buffer) {
	fmt.Printf("%+v\n", buf)
}

func testRingBufGrow() {
	fmt.Println("==== testRingBufGrow ====")

	// r == w == 0
	buf := NewBuffer(1)
	buf.Write([]byte{1,2})
	printfBuf(buf)

	// w > r 
	fmt.Println("Read:", buf.Read(buf.Len()/2))
	buf.Write([]byte{3,4,5})
	printfBuf(buf)

	// r == w == cap
	fmt.Println("Read:", buf.Read(buf.Len()))
	buf.Write([]byte{6,7,8,9})
	printfBuf(buf)
	buf.Write([]byte{10,11,12,13})
	printfBuf(buf)

	// w < r
	fmt.Println("Read:", buf.Read(buf.Len()/2))
	buf.Write([]byte{14,15,16})
	printfBuf(buf)
	buf.Write([]byte{17,18,19,20})
	printfBuf(buf)

	// r == w == mid
	buf1 := NewBuffer(4)
	buf1.Write([]byte{1,2,3,4})
	printfBuf(buf1)
	fmt.Println("Read:", buf1.Read(buf1.Len()/2))
	buf1.Write([]byte{5,6})
	printfBuf(buf1)
	buf1.Write([]byte{10,11,12,13})
	printfBuf(buf1)
}

func testRingBufWrite() {
	fmt.Println("==== testRingBufWrite ====")

	// w == r && w == 0
	buf := NewBuffer(8)
	printfBuf(buf)
	buf.Write([]byte{1,2})
	printfBuf(buf)
	
	// w == r && w != 0
	fmt.Println("Read:", buf.Read(buf.Len()))
	buf.Write([]byte{3,4})
	printfBuf(buf)

	// w == r && w == cap
	buf.Write([]byte{5,6,7,8})
	fmt.Println("Read:", buf.Read(buf.Len()))
	printfBuf(buf)
	buf.Write([]byte{10,11})
	printfBuf(buf)

	// w < r
	buf.Write([]byte{12,13})
	printfBuf(buf)

	// w > r && cap-w is enough
	fmt.Println("Read:", buf.Read(buf.Len()/2))
	printfBuf(buf)
	buf.Write([]byte{14,15})
	printfBuf(buf)

	// w > r && cap-w is not enough
	buf.Write([]byte{16,17,18})
	printfBuf(buf)
}

func testRingBufRead() {
	fmt.Println("==== testRingBufRead ====")

	// w == r && w == 0
	buf := NewBuffer(8)
	fmt.Println("Read:", buf.Read(buf.Len()))
	printfBuf(buf)

	// w > r 
	buf.Write([]byte{1,2})
	fmt.Println("Read:", buf.Read(buf.Len()))
	printfBuf(buf)

	// w == r && w != 0 && cap-r is enough
	buf.Write([]byte{3,4,5,6,7,8,9,10})
	printfBuf(buf)
	fmt.Println("Read:", buf.Read(2))
	printfBuf(buf)

	// w == r && w != 0 && cap-r is not enough
	buf.Write([]byte{11,12})
	printfBuf(buf)
	fmt.Println("Read:", buf.Read(buf.Len()))
	printfBuf(buf)

	// w < r && cap-r is enough
	buf.Write([]byte{21,22,23,24,25})
	fmt.Println("Read:", buf.Read(2))
	printfBuf(buf)

	// w < r && cap-r is not enough
	fmt.Println("Read:", buf.Read(buf.Len()))
	printfBuf(buf)

}

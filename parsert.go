package main

import (	
	"fmt"
    "os"
	"log"	
    "regexp"
	"time"
)

var html_id int;

var html_len int;

var emptyStartTag=regexp.MustCompile(`<\s*(?P<tag>\w{1,7})\s*>`);
var startTag=regexp.MustCompile(`<\s*(?P<tag>\w{1,7})\s+(([\w\s\d]+(\=[\"\'][^\"\']*[\"\'])?)+)\s*>`);
var endTag=regexp.MustCompile(`<!--[^<>\/]*-->`);

var debug *log.Logger;

var ts time.Time;
var td time.Duration;

type ParsedNode struct{
	start,end int;
	tag string;
	nid string;
	class string;
	level int;
	styles *map[string]string;
	content string;
	parents *[]ParsedNode;
	index int;
	nextSibling *ParsedNode;
	prevSibling *ParsedNode;
	path string;
}

func testErr(err error) bool{
	if err!=nil {
		log.Fatal(err);
		return false;
	}
	return true;
}

func analyzeImplement(n int) func(html *[]byte) (*[][]byte,*[][]byte,*[][]byte){
	count:=n;
	//var curPos int;
	//curPos=0;

	return func(html *[]byte) (*[][]byte,*[][]byte,*[][]byte){
		fmt.Println(startTag);
		//bys:=[]byte(*html);
		etmps:=emptyStartTag.FindAll(*html,count);
		tmps:=startTag.FindAll(*html,count);
		tmpe:=endTag.FindAll(*html,count);
		return &etmps,&tmps,&tmpe;
	}
}

func analyzeHtml(html *[]byte){
	ts=time.Now();
	fmt.Println("yes, it works!");
	an:=analyzeImplement(10000);
	estart,start,end:=an(html);
	n0:=len(*estart);
	n:=len(*start);
	m:=len(*end);
	debug.Printf("\n*******************\ntag without attribute\n**********************\n");
	for i:=0;i<n0;i++{
		debug.Println(string((*estart)[i]));
	}
	debug.Printf("\n*******************\ntag with attribute\n**********************\n");
	for i:=0;i<n;i++{
		debug.Println(string((*start)[i]));
	}
	debug.Printf("\n*******************\nend tag \n**********************\n");
	for i:=0;i<m;i++{
		debug.Println(string((*end)[i]));
	}
	td=ts.Sub(time.Now());
	fmt.Printf("emptystart:%d\tstart:%d\tend:%d\ntime spent:%s",n0,n,m,td.String());
}

func doTheJob(){
	data:=make([]byte,512*1024*1024,1024*1024*1024);
	file,err:=os.Open("test.htm");
	if !testErr(err){
	
	}
	count , err := file.Read(data);
	html_len=count;
	if !testErr(err){
	
	}
	//n:=count/1000;
	//fmt.Printf("read %d bytes: %q\n", count, data[:n]);
	//html:=string(data);
	analyzeHtml(&data);
}

func main(){
     fmt.Println("handling file");
	 file,err:=os.Create("debug.txt");
	 if !testErr(err){
		
	 }
	 debug=log.New(file,"debug:\t",4);
	 doTheJob();
}
package main

import (	
	"fmt"
    "os"
	"log"	
    "regexp"
	"time"
)

var html_id int;
var lsNodeId int;
var emptyStartTag=regexp.MustCompile(`<\s*(?P<tag>[a-zA-Z1-6]{1,7})\s*>\s*`);
var startTag=regexp.MustCompile(`<\s*(?P<tag>[a-zA-Z1-6]{1,7})((?:\s+[\w\-]+(?:\=[\"\'][^\"\']*[\"\'])?)+)\s*>\s*`);
var endTag=regexp.MustCompile(`<\s*/\s*(?P<tag>[a-zA-Z1-6]{1,7})\s*>\s*`);
var attributes=regexp.MustCompile(`\s+([\w\-]+)(?:\=([\"\'][^\"\']*[\"\']))?`);
//var emptyStr=regexp.MustCompile(`\s+`);
var analyzedAttrs=map[string]string{
	"id":"id"
	"class":"class"
	"style":"style"
	"bgcolor":"style"
	"align":"style"	
}
var debug *log.Logger;
var ts time.Time;
var td time.Duration;
var results *[]ParsedNode;

type Positions struct{
	start,end int;
}

type ParsedNode struct{
	nodeId int;
	strPos,endPos *Positions;	
	tag,nid,path,content string;	
	classes *[]string;
	level,index,childCount int16;
	styles,otherAttrs *map[string]string;	
	parents *[]*ParsedNode;
		
}

type NodeStack struct{
	stack []*ParsedNode;
	count int16;
}
func (stack *NodeStack) init(c int){
	stack.stack=make([]*ParsedNode,0,c);
	stack.count=0;
}
func (stack *NodeStack) pushEl(node *ParsedNode){
	stack.stack[count]=node;
	stack.count++;
}
func (stack *NodeStack) popEl() *ParsedNode{
	stack.count--;
	re:=stack.stack[count];
	stack.stack[count]=nil;
	return re;
}
func (stack *NodeStack) top() *ParsedNode{
	if stack.count==0{
		return nil;
	}else{
		return stack.stack[count-1];
	}
}

func (node *ParsedNode) string() string{
	classes:="";
	for _,v := range *node.classes {
		classes+=v+".";
	}
	return node.tag+"#"+node.nid+"."+classes+"\tpos:"+string(node.level)+","+string(node.index);
}

func testErr(err error) bool{
	if err!=nil {
		log.Fatal(err);
		return false;
	}
	return true;
}
/**
func nnlen(slice []*ParsedNode) int{
	n:=0;
	for _,v := range slice {
		if v!=nil {
			n++;
		}
	}
	return n;
}
**/
func analyzeImplement(data *[]byte,offset,length int) func(start int) (int,bool){
	lCount:=0;
	//index:=0;
	//level:=0;
	html:=(*data)[offset:];
	nodeStack:=new(NodeStack);
	nodeStack.init(70);
	styleStack:=make([]*map[string]string,17);
	lens:=length;
	//firstChild:=true;
	var nextStart *[]int;
	nextStart=nil;
	res:=make([]ParsedNode,0,5000);
	
	AttributeHandle:=func(attrs string) (string,*[]string,*map[string]string,*map[string]string){
		re:=attributes.FindAllStringSubmatch(attrs,15);
		m:=len(re);	
		nid:="";
		classes=make([]string,7);
		styles=make(map[string]string);
		attrs=make(map[string]string);
		
		//copy all parents' style
		ns:=len(styleStack);
		for _,sts range styleStack {
			
		}
		
		for j:=0;j<m;j++ {
			mm:=len(re[j]);
			if mm==3{
				k:=re[j][1];
				tp,ok:=analyzedAttrs[k];
				if ok {
					switch tp {
						case "id" :
							nid=re[j][2];
						case "class":
							append(classes,re[j][2]);
						case "styles":
							styles[""]
					}
				} else{
				}
			}
		}
		
		return nid,&classes,&styles,&attrs;
	}
		
	//strPos,nodeId,tag,nid,class,styles
	NodeStartTagHandle:=func(pos int[],tag,attrs string){
		//get positions
		p:=new(Positions);
		p.start,pos.end=pos[0],pos[1];
		
		//build node
		node:=new(ParsedNode);
		
		//set tag
		node.tag=tag;
		
		//set start position
		node.strPos=pos;
		
		//set nodeId
		lsNodeId++;
		node.nodeId=lsNodeId;
		
		//analyze attrs
		if attrs!=""{
			node.nid="";
			node.classes=nil;
			node.styles=nil;
			node.otherAttrs=nil;
		}else{			
			node.nid,node.classes,node.styles,node.otherAttrs=AttributeHandle(attrs);
		}
		
		nodeStack.pushEl(node);		
	}
	//endPos,content,level,index,(childCount),parents 
	NodeEndTagHandle:=func(pos int[],tag string){
		//get positions
		p:=new(Positions);
		p.start,pos.end=pos[0],pos[1];
		
		//set parents
		n:=nodeStack.count;
		prs:=make([]*ParsedNode,n);
		copy(nodeStack.stack[:n],prs);
		node.parents=&prs;
		
		//set level
		node.level = int16(n);
		
				
		//add child count of parent node 
		pNode:=nodeStack.top();
		pNode.childCount++;
		
	}
	
	//node need properties : 
	// properties to be setted at start: strPos,nodeId,tag,nid,class,styles
	// properties to be setted at end  : endPos,content,level,index,(childCount),parents 
	return func(start int) (int,bool){
		if start>=lens {	
			results=&res;		
			return -1,false;
		}
			
		// try to find start tag from [start] position
		if nextStart==nil {
			flag:=true;
			tmp:=emptyStartTag.FindIndex(html[start:]);
			if tmp==nil{
				tmp=startTag.FindIndex(html[start:]);
				flag=false;
			}
		} else {
			tmp=*nextStart;
		}
		
		if tmp!=nil{
			//if no other content between [start] position and this tag
			if tmp[0]==start{
				//handle start tag
				if flag {
					//build node 
					dt:=emptyStartTag.FindSubmatch(html[start:]);
					NodeStartTagHandle(tmp,string(dt[1]),"");
				}else{
					//build node 
					dt:=startTag.FindSubmatch(html[start:]);
					NodeStartTagHandle(tmp,string(dt[1]),string(dt[2]));
				}
												
				return tmp[1]+1,true;
			} else{
				nextStart=&tmp;
			}
		} 
			
		//try to find end tag
		if tmp:=endTag.FindIndex(html[start:]);tmp!=nil{
			//handle end tag
			dt:=endTag.FindSubmatch(html[start:]);
			NodeEndTagHandle(tmp,string(dt[1]));
			return tmp[1]+1,true;
		} 
		
		//neither start tag nor end tag found , end parsing and return the parsed length		
		return start,false;
	}
}

func analyzeHtml(){
	ts=time.Now();
	data:=make([]byte,0,1024*1024*1024);
	file,err:=os.Open("test.htm");
	if !testErr(err){
	
	}
	len , err := file.Read(data);
	//html_len=len;
	if !testErr(err){
	
	}
	//n:=count/1000;
	fmt.Printf("read %d \n", len);
	
	fmt.Println("yes, it is working!");
	analyzer:=analyzeImplement(&data,0,len);
	start:=0;
	finish:=false;
	for {
		if !finish{
			start,finish=analyzer(start);
		}else{
			break;
		}
		
	}
	if start==-1{
		fmt.Println("analyzed whole file");
	}else{
		fmt.Println("analyzed "+start+" bytes of file");
	}
		
	td=ts.Sub(time.Now());
	fmt.Printf("\ntime spent:%s\n",td.String());
}

/**
func doTheJob(){
	
	//html:=string(data);
	analyzeHtml(&data);
}
**/

func main(){
     fmt.Println("handling file");
	 file,err:=os.Create("debug.txt");
	 if !testErr(err){
		
	 }
	 debug=log.New(file,"debug:\t",4);
	 analyzeHtml();
}

/**
a = append(a, b...)
Copy

b = make([]T, len(a))
copy(b, a)
Cut

a = append(a[:i], a[j:]...)
Delete

a = append(a[:i], a[i+1:]...)
Expand

a = append(a[:i], append(make([]T, j), a[i:]...)...)
Extend

a = append(a, make([]T, j)...)
Insert

a = append(a[:i], append([]T{x}, a[i:]...)...)
InsertVector

a = append(a[:i], append(b, a[i:]...)...)
Pop

x, a = a[len(a)-1], a[:len(a)-1]
Push

a = append(a, x)
**/
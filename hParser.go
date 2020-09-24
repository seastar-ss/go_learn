package main

import (	
	"fmt"
    "os"
	"log"	
    "regexp"
	"time"
	"strings"
)

var html_id int;
var lsNodeId int;
var scriptTag=regexp.MustCompile(`^\s*<\s*script.*>(?:.|\r|\n)*?</\s*script\s*>`);
var emptyStartTag=regexp.MustCompile(`^\s*<\s*(?P<tag>[a-zA-Z1-6\:]{1,9})\s*\/?\s*>`);
var startTag=regexp.MustCompile(`^\s*<\s*(?P<tag>[a-zA-Z1-6\:]{1,9})((?:\s+[\w\-]+?(?:\=\"[^\"]*\")?)+)\s*\/?\s*>`);
var textContent=regexp.MustCompile(`^[^<>]+`);
var endTag=regexp.MustCompile(`^\s*<\s*/\s*(?P<tag>[a-zA-Z1-6\:]{1,9})\s*>`);
var comment=regexp.MustCompile(`^\s*<!--([^\-]+-?)*-->`);
var attributes=regexp.MustCompile(`\s+([\w\-]+)(?:\=[\"\']([^\"\']*)[\"\'])?`);
var stylesAttr=regexp.MustCompile(`\s*([\w\-]+)\s*\:\s*([^;\:]+)\s*;`);
//var emptyStr=regexp.MustCompile(`\s+`);
var analyzedAttrs=map[string]string{
	"id":"id",
	"class":"class",
	"style":"style",
	"bgcolor":"style",
	"align":"style",	
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
	parent *ParsedNode;
	children *[]*ParsedNode;	
}

func (node *ParsedNode) copy(cpNode *ParsedNode){
	node.nodeId,node.strPos,node.endPos,node.tag,node.nid,node.path,node.content=cpNode.nodeId,cpNode.strPos,cpNode.endPos,cpNode.tag,cpNode.nid,cpNode.path,cpNode.content;
	node.classes,node.level,node.index,node.childCount,node.styles,node.otherAttrs,node.parent,node.children=cpNode.classes,cpNode.level,cpNode.index,cpNode.childCount,cpNode.styles,cpNode.otherAttrs,cpNode.parent,cpNode.children;
}

func (node *ParsedNode) string() string{
	if node==nil{
		return "nil";
	}
	classes:="";
	if node.classes!=nil{
		for _,v := range *node.classes {
			classes+=v+".";
		}
	}
	return node.tag+"#"+node.nid+"."+classes+"\tpos:"+string(node.level)+","+string(node.index)+"\t--"+node.content;
}

func (node *ParsedNode) init(){
	node.strPos,node.endPos,node.parent,node.children=nil,nil,nil,nil;
	node.classes,node.styles,node.otherAttrs=nil,nil,nil;
	node.level,node.index=0,0;
	node.content,node.tag="","";
}

type NodeStack struct{
	stack []*ParsedNode;
	count,cp int16;
	
}
func (stack *NodeStack) init(c int){
	stack.stack=make([]*ParsedNode,c);
	stack.count=0;
	stack.cp=int16(c);
}
func (stack *NodeStack) pushEl(node *ParsedNode){
	if stack.count>=stack.cp{
		tmp:=make([]*ParsedNode,stack.cp+50);
		copy(tmp,stack.stack);
		stack.stack=tmp;
	}
	stack.stack[stack.count]=node;
	stack.count++;
}
func (stack *NodeStack) popEl() *ParsedNode{
	stack.count--;
	re:=new(ParsedNode);
	re.copy(stack.stack[stack.count]);
	stack.stack[stack.count]=nil;
	return re;
}
func (stack *NodeStack) top() *ParsedNode{
	if stack.count!=0{		
		return stack.stack[stack.count-1];
	}
	return nil;
}

func testErr(err error) bool{
	if err!=nil {
		debug.Println(err);
		return false;
	}
	return true;
}

func analyzeImplement(data *[]byte,offset,length int) func(start int) (int,bool){
	//lCount:=0;
	//index:=0;
	//level:=0;
	html:=(*data)[offset:];
	nodeStack:=new(NodeStack);
	nodeStack.init(177);
	//styleStack:=make([]*map[string]string,17);
	lens:=length;
	//firstChild:=true;
	//var nextStart *[]int;
	//nextStart=nil;
	res:=make([]ParsedNode,4500,5000);
	//texts:=make([]string,30);
	
	AttributeHandle:=func(attrs string) (string,*[]string,*map[string]string,*map[string]string){
		re:=attributes.FindAllStringSubmatch(attrs,17);
		m:=len(re);	
		nid:="";
		classes:=make([]string,7);
		styles:=make(map[string]string);
		oAttrs:=make(map[string]string);
				
		//analyze the attribute key/value
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
							classes=strings.Fields(re[j][2]);
						case "styles":
							if re[j][1]!="style"{
								styles[re[j][1]]=re[j][2];
							}else{
								stys:=stylesAttr.FindAllStringSubmatch(re[j][2],17);
								for _,vv:=range stys{
									if len(vv)==3{
										styles[vv[1]]=vv[2];
									}
								}
							}

					}
				} else{
					oAttrs[re[j][1]]=re[j][2];
				}
			}
		}
		
		return nid,&classes,&styles,&oAttrs;
	}
		
	//strPos,nodeId,tag,nid,class,(styles),otherAttr
	NodeStartTagHandle:=func(pos []int,tag,attrs string){
		//get positions
		p:=new(Positions);
		p.start,p.end=pos[0],pos[1];
		
		//build node
		node:=new(ParsedNode);
		node.init();
		//set tag
		node.tag=tag;
		
		//set start position
		node.strPos=p;
		
		//set nodeId
		lsNodeId++;
		node.nodeId=lsNodeId;
		
		//analyze attrs
		if attrs!=""{
			node.nid,node.classes,node.styles,node.otherAttrs="",nil,nil,nil;			
		}else{			
			node.nid,node.classes,node.styles,node.otherAttrs=AttributeHandle(attrs);
		}
		
		nodeStack.pushEl(node);		
	}
	//endPos,content,(level),index,childCount,children,parent 
	NodeEndTagHandle:=func(pos []int,tag string){
		//get positions
		p:=new(Positions);
		p.start,p.end=pos[0],pos[1];
		
		counts:=0;
				
		var node *ParsedNode;
		
		children:=make([]*ParsedNode,0,77);
		
		for node=nodeStack.top();node.tag!=tag && nodeStack.count>0;node=nodeStack.top(){			
			node.index=int16(-counts);
			counts++;
			children=append(children,nodeStack.popEl());
		}
		fmt.Println(children);
		debug.Println(node.string());
		for _,ch:=range children{
			debug.Println(ch.string());
			ch.index+=int16(counts);
			ch.parent=node;
			//node.content+=ch.content+"_~~~_";
		}
		node.children=&children;
		node.childCount=int16(counts);
		node.endPos=p;

	}
	
	TextHandle:=func (dt string,pos []int) {
		tx:=strings.Trim(dt," ");
		node:=new(ParsedNode);
		node.init();
		
		ps:=new(Positions);
		ps.start,ps.end=pos[0],pos[0];
		
		pe:=new(Positions);
		pe.start,pe.end=pos[1],pos[1];
		
		node.strPos,node.endPos=ps,pe;
		
		node.content=tx;
		
		lsNodeId++;
		node.nodeId=lsNodeId;
		
		node.tag,node.nid="text","";
		
		node.classes,node.styles=nil,nil;
		
		node.childCount=0;
		
		nodeStack.pushEl(node);
	}
	
	GenResult:=func() *[]ParsedNode{
		return &res;
	}
	
	//node need properties : 
	// properties to be setted at start: strPos,nodeId,tag,nid,class,styles
	// properties to be setted at end  : endPos,content,level,index,(childCount),parents 
	return func(start int) (int,bool){
		if start>=lens {	
			results=GenResult();		
			return 0,false;
		}
		fmt.Printf("start:%d\n",start);
		var re int;
		dd:=html[start:];
		//debug.Println(string(dd[0:30]));
		var tmp []int;
		flag:=false;
		tmp=scriptTag.FindIndex(dd);
		if tmp!=nil{
			//handle script or not
			tmp[0],tmp[1]=tmp[0]+start,tmp[1]+start;
			flag=true;
		}
		// try to find start tag from [start] position	
		if !flag {
			tmp=emptyStartTag.FindIndex(dd);
			if tmp!=nil{					
				//handle start tag		
				//build node no attr
				dt:=emptyStartTag.FindSubmatch(dd[:tmp[1]]);
				tmp[0],tmp[1]=tmp[0]+start,tmp[1]+start;
				NodeStartTagHandle(tmp,string(dt[1]),"");
				flag=true;
				//return tmp[1],true;			
			}
		}
		if !flag{
			tmp=startTag.FindIndex(dd);
			if tmp!=nil{		
				//handle start tag			
				//build node with attr
				dt:=startTag.FindSubmatch(dd[:tmp[1]]);
				tmp[0],tmp[1]=tmp[0]+start,tmp[1]+start;
				NodeStartTagHandle(tmp,string(dt[1]),string(dt[2]));
				flag=true;			
				//return tmp[1],true;
			}
		}
		if !flag{
			tmp=textContent.FindIndex(dd);
			if tmp!=nil{
				//handle text
				dt:=dd[tmp[0]:tmp[1]];
				tmp[0],tmp[1]=tmp[0]+start,tmp[1]+start;
				TextHandle(string(dt),tmp);
				flag=true;
				//return tmp[1],true;
			}
		}	
		//try to find end tag
		if !flag{
			tmp=endTag.FindIndex(dd);
			if tmp!=nil{
				//handle end tag
				dt:=endTag.FindSubmatch(dd[:tmp[1]]);
				tmp[0],tmp[1]=tmp[0]+start,tmp[1]+start;
				NodeEndTagHandle(tmp,string(dt[1]));
				flag=true;
				//return tmp[1],true;
			} 
		}
		//try to find comment
		if !flag{
			tmp=comment.FindIndex(dd);
			if tmp!=nil{
				//handle comment or not
				tmp[0],tmp[1]=tmp[0]+start,tmp[1]+start;
				flag=true;
			}
		}
		if flag{
			debug.Println("matched:"+string(dd[:tmp[1]-start]));
			re=tmp[1];
		}else{
			debug.Println("wtf:"+string(dd[0:30]));
			re=start;
		}
		//neither start tag nor end tag found , end parsing and return the parsed length	
		
		return re,flag;
	}
}

func analyzeHtml(){
	ts=time.Now();
	data:=make([]byte,512*1024*1024,1024*1024*1024);
	file,err:=os.Open("test.htm");
	if !testErr(err){
	
	}
	lens , err := file.Read(data);
	//html_len=len;
	if !testErr(err){
	
	}
	//n:=count/1000;
	fmt.Printf("read %d \n", lens);
	fmt.Println(string(data));
	fmt.Println("yes, it is working!");
	analyzer:=analyzeImplement(&data,0,lens);
	start:=0;
	conti:=true;
	for {
		if conti{
			start,conti=analyzer(start);
			fmt.Println(start,conti);
		}else{
			break;
		}
	}
	if start==0{
		fmt.Println("analyzed whole file");
	}else{
		fmt.Printf("analyzed %d bytes of file\n",start);
	}
		
	td=ts.Sub(time.Now());
	fmt.Printf("\ntime spent:%s\n",td.String());
}

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
/**
func doTheJob(){
	
	//html:=string(data);
	analyzeHtml(&data);
}
**/


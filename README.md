simplelang
==========
This is an attempt at writing a basic compiler for self-study purposes.

The source language will be PL/0. The target language is MIPS assembly.
The implementation language is Go.

The grammar for the language is mostly ripped from the Wikipedia page on PL/0. The ? operator was
removed. My usage of it is defined in EBNF form as follows:
```
program = block "." .

block = [ "const" ident "=" number {"," ident "=" number} ";"]
        [ "var" ident {"," ident} ";"]
        { "procedure" ident ";" block ";" } statement .

statement = [ ident ":=" expression | "call" ident 
              | "!" expression 
              | "begin" statement {";" statement } "end" 
              | "if" condition "then" statement 
              | "while" condition "do" statement ]

condition = "odd" expression |
            expression ("="|"#"|"<"|"<="|">"|">=") expression .

expression = [ "+"|"-"] term { ("+"|"-") term}.

term = factor {("*"|"/") factor}.

factor = ident | number | "(" expression ")".
```

### Usage
If you run go install and have $GOPATH set up, run `simplelang FILE`
The compiler will try to interpret any PL/0 code. If there are syntax errors, the compiler will only
print the line number of the first one. If there are semantic errors, the compiler will list all of
them. If the compilation is successful, an output file out.s will be produced with SPIM assembly.
Use QtSpim or command line Spim to run it.

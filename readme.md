
# Glox

A Go implementaion of the tree walk interpreter for the Lox language. I wrote this version
while reading the book [Crafting Interpreters](http://craftinginterpreters.com/)
by Robert Nystrom.

## Documentation

### Installation
Download the latest binary from the [Releases](https://github.com/iamsayantan/glox/releases) page
as per your platform. Unzip the package, you'll have the glox binary. Running only `./glox`
will give you the interactive terminal or you can pass in the location of a glox script file to 
run a script e.g. `./glox hello.glox` where `hello.glox` contains the glox script in the same
directory as the glox binary.

### Examples

#### Hello world: 
```
// Your very first lox program.
print "Hello, world";
```

#### Variable declaration
```
var a = 5;
var b = 6;

var c = a + b;
print c; // prints 11
```

#### Loops
```
for (var i = 0; i < 5; i = i+1) {
  print i;
}

// prints 
// 0
// 1
// 2
// 3
// 4
```

#### Conditionals
```
var a = 5;
if (a >= 4) {
  print "a is greater than or equal to 4";
} else {
  print "a is less than 4";
}
```
#### Functions
```
fun printSum(a, b) {
  print a + b;
}

printSum(5, 6); // prints 11

// Closure
fun returnFunction() {
  var outside = "outside";

  fun inner() {
    print outside;
  }

  return inner;
}

var fn = returnFunction();
fn(); // prints outside
```
#### Classes
```
class Breakfast {
  // constructor
  init(meat, bread) {
    this.meat = meat;
    this.bread = bread;
  }

  cook() {
    print "Eggs a-fryin'!";
  }

  serve(who) {
    print "Enjoy your " + this.meat + " and " +
        this.bread + ", " + who + ".";
  }
}

// instantiate
var breakfast = Breakfast("sausage", "toast");
breakfast.serve("Sayantan"); // prints "Enjoy your sausage and toast, Sayantan".
```
#### Inheritence
```
class Brunch < Breakfast {
  init(meat, bread, drink) {
    super.init(meat, bread);
    this.drink = drink;
  }
}
```
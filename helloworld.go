package main
import ("fmt")

type addr struct{
	street string
	city string
}
type Person struct{
	name string
	age int
	gender string
	address addr
}
func PersonConstructor(name string,age int,gender string,address addr) *Person{
	p := Person{
		name: name,
		age: age,
		gender: gender,
		address: addr{
			street: address.street,
			city: address.city,
		},
	}
	return &p

}

func main() {
	p1 := PersonConstructor("tom",18,"male",addr{"123main","nj"})
	fmt.Println(p1.name)
	fmt.Println(p1.age)
	fmt.Println(p1.gender)
	fmt.Println(p1.address.street,p1.address.city)
}
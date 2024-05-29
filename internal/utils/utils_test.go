package utils

import (
	"fmt"
	"testing"
)

func TestCheckOrderByLuna(t *testing.T) {

	tests := []struct {
		name     string
		orderNum string
		want     bool
	}{
		// TODO: Add test cases.
		{
			name:     "LunaTest valid order",
			orderNum: "79927398713",
			want:     true,
		},
		{
			name:     "LunaTest invalid order",
			orderNum: "09927398713",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckOrderByLuna(tt.orderNum); got != tt.want {
				t.Errorf("CheckOrderByLuna() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {

	tests := []struct {
		name        string
		tokenString string
		userID      int
		want        bool
	}{
		// TODO: Add test cases.
		{
			name:        "Get UserID from auth token",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI3MzgzOTAsIlVzZXJJRCI6NH0.V9WdWdJWeU1qqVCGDfTGu0asPZhiFUPmtnsfpN0GPro",
			userID:      4,
			want:        true,
		},
		{
			name:        "Get UserID from auth token ",
			tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI3MzgzOTAsIlVzZXJJRCI6NH0.V9WdWdJWeU1qqVCGDfTGu0asPZhiFUPmtnsfpN0GPro",
			userID:      5, // it can any int for test, excluding 4
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := GetUserID(tt.tokenString)
			if (userID == tt.userID) != tt.want {
				t.Errorf("GetUserID() = %v, compare with = %v,  want %v", userID, tt.userID, tt.want)
			}
		})
	}
}

func ExampleGetUserID() {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI3MzgzOTAsIlVzZXJJRCI6NH0.V9WdWdJWeU1qqVCGDfTGu0asPZhiFUPmtnsfpN0GPro"
	//userID := 5
	out1 := GetUserID(tokenString)
	fmt.Println(out1)

	//Output
	//5

}

func ExampleCheckOrderByLuna() {
	order1 := "1004128237584"
	out1 := CheckOrderByLuna(order1)
	fmt.Println(out1)

	order2 := "100412823758"
	out2 := CheckOrderByLuna(order2)
	fmt.Println(out2)

	//Output
	//True
	//Fase
}

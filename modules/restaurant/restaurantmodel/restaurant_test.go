package restaurantmodel

import (
	"testing"
)

type DataTable struct {
	Input  RestaurantCreate
	Expect error
}

func TestValidate(t *testing.T) {

	table := []DataTable{
		{Input: RestaurantCreate{Name: ""}, Expect: ErrNameCannotBeEmpty},
		{Input: RestaurantCreate{Name: "i have title"}, Expect: nil},
	}

	for i, v := range table {
		err := table[i].Input.Validate()

		if err != v.Expect {
			t.Errorf("Test Validate() failed, expected %v, but got %v ", v.Expect, err)
		}
	}

	t.Log("Test Validate() pass")
}

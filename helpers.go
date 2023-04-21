package main

import (
	"strconv"
	"strings"
)

// func newReg() string {
// 	return strings.Replace(uuid.New().String(), "-", "", -1)
// }

// var id int32 = 999

// func addTenant(tenant *Tenant) {

// 	id++
// 	tenant.Id = id
// 	tenant.Reg = strings.Replace(uuid.New().String(), "-", "", -1)

// 	tenants = append(tenants, *tenant)
// }

func isValidCpf(cpf string) (bool, string) {

	_, err := strconv.ParseInt(cpf, 10, 64)
	if err != nil {
		return false, "Cpf inválido"
	}

	for i := 0; i < 10; i++ {
		if cpf == strings.Repeat(strconv.Itoa(i), 11) {
			return false, "Cpf inválido"
		}
	}

	// fmt.Println(strings.Repeat("d", 14))
	if len(cpf) != 11 {
		return false, "O cpf dever conter 11 digitos"
	}

	// Primeiro digito
	var add int32 = 0
	digs := strings.Split(cpf, "")

	for i := 0; i < 9; i++ {
		dig, _ := strconv.ParseInt(digs[i], 10, 8)
		add += int32(dig)
	}

	// Segundo digito
	add = 0
	for i := 0; i < 10; i++ {
		dig, _ := strconv.ParseInt(digs[i], 10, 8)
		add += int32(dig) * (11 - int32(i))
	}

	rev := 11 - (add % 11)
	if rev == 10 || rev == 11 {
		rev = 0
	}

	digVerificador, _ := strconv.ParseInt(digs[10], 10, 8)
	if rev != int32(digVerificador) {
		return false, "Cpf inválido"
	}

	return true, ""
}

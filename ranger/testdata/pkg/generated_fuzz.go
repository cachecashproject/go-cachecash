package pkg

import (
	"fmt"
	"reflect"
)

func Fuzz(data []byte) int {
	if len(data) == 0 {
		/* need a type signature byte */
		return -1
	}

	switch int(data[0]) % 14 {

	case 0:
		objEscrowOpenTransaction := &EscrowOpenTransaction{}
		_, err := objEscrowOpenTransaction.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataEscrowOpenTransaction, err := objEscrowOpenTransaction.Marshal()
		if err != nil {
			panic(err)
		}

		objEscrowOpenTransaction2 := &EscrowOpenTransaction{}
		_, err = objEscrowOpenTransaction2.UnmarshalFrom(dataEscrowOpenTransaction)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objEscrowOpenTransaction, objEscrowOpenTransaction2) {
			panic(fmt.Sprintf("obj %T not equal", objEscrowOpenTransaction))
		}

	case 1:
		objGenesisTransaction := &GenesisTransaction{}
		_, err := objGenesisTransaction.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataGenesisTransaction, err := objGenesisTransaction.Marshal()
		if err != nil {
			panic(err)
		}

		objGenesisTransaction2 := &GenesisTransaction{}
		_, err = objGenesisTransaction2.UnmarshalFrom(dataGenesisTransaction)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objGenesisTransaction, objGenesisTransaction2) {
			panic(fmt.Sprintf("obj %T not equal", objGenesisTransaction))
		}

	case 2:
		objGlobalConfigListInsertion := &GlobalConfigListInsertion{}
		_, err := objGlobalConfigListInsertion.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataGlobalConfigListInsertion, err := objGlobalConfigListInsertion.Marshal()
		if err != nil {
			panic(err)
		}

		objGlobalConfigListInsertion2 := &GlobalConfigListInsertion{}
		_, err = objGlobalConfigListInsertion2.UnmarshalFrom(dataGlobalConfigListInsertion)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objGlobalConfigListInsertion, objGlobalConfigListInsertion2) {
			panic(fmt.Sprintf("obj %T not equal", objGlobalConfigListInsertion))
		}

	case 3:
		objGlobalConfigListUpdate := &GlobalConfigListUpdate{}
		_, err := objGlobalConfigListUpdate.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataGlobalConfigListUpdate, err := objGlobalConfigListUpdate.Marshal()
		if err != nil {
			panic(err)
		}

		objGlobalConfigListUpdate2 := &GlobalConfigListUpdate{}
		_, err = objGlobalConfigListUpdate2.UnmarshalFrom(dataGlobalConfigListUpdate)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objGlobalConfigListUpdate, objGlobalConfigListUpdate2) {
			panic(fmt.Sprintf("obj %T not equal", objGlobalConfigListUpdate))
		}

	case 4:
		objGlobalConfigScalarUpdate := &GlobalConfigScalarUpdate{}
		_, err := objGlobalConfigScalarUpdate.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataGlobalConfigScalarUpdate, err := objGlobalConfigScalarUpdate.Marshal()
		if err != nil {
			panic(err)
		}

		objGlobalConfigScalarUpdate2 := &GlobalConfigScalarUpdate{}
		_, err = objGlobalConfigScalarUpdate2.UnmarshalFrom(dataGlobalConfigScalarUpdate)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objGlobalConfigScalarUpdate, objGlobalConfigScalarUpdate2) {
			panic(fmt.Sprintf("obj %T not equal", objGlobalConfigScalarUpdate))
		}

	case 5:
		objGlobalConfigTransaction := &GlobalConfigTransaction{}
		_, err := objGlobalConfigTransaction.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataGlobalConfigTransaction, err := objGlobalConfigTransaction.Marshal()
		if err != nil {
			panic(err)
		}

		objGlobalConfigTransaction2 := &GlobalConfigTransaction{}
		_, err = objGlobalConfigTransaction2.UnmarshalFrom(dataGlobalConfigTransaction)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objGlobalConfigTransaction, objGlobalConfigTransaction2) {
			panic(fmt.Sprintf("obj %T not equal", objGlobalConfigTransaction))
		}

	case 6:
		objOutpoint := &Outpoint{}
		_, err := objOutpoint.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataOutpoint, err := objOutpoint.Marshal()
		if err != nil {
			panic(err)
		}

		objOutpoint2 := &Outpoint{}
		_, err = objOutpoint2.UnmarshalFrom(dataOutpoint)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objOutpoint, objOutpoint2) {
			panic(fmt.Sprintf("obj %T not equal", objOutpoint))
		}

	case 7:
		objTransaction := &Transaction{}
		_, err := objTransaction.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataTransaction, err := objTransaction.Marshal()
		if err != nil {
			panic(err)
		}

		objTransaction2 := &Transaction{}
		_, err = objTransaction2.UnmarshalFrom(dataTransaction)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objTransaction, objTransaction2) {
			panic(fmt.Sprintf("obj %T not equal", objTransaction))
		}

	case 8:
		objTransactionInput := &TransactionInput{}
		_, err := objTransactionInput.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataTransactionInput, err := objTransactionInput.Marshal()
		if err != nil {
			panic(err)
		}

		objTransactionInput2 := &TransactionInput{}
		_, err = objTransactionInput2.UnmarshalFrom(dataTransactionInput)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objTransactionInput, objTransactionInput2) {
			panic(fmt.Sprintf("obj %T not equal", objTransactionInput))
		}

	case 9:
		objTransactionOutput := &TransactionOutput{}
		_, err := objTransactionOutput.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataTransactionOutput, err := objTransactionOutput.Marshal()
		if err != nil {
			panic(err)
		}

		objTransactionOutput2 := &TransactionOutput{}
		_, err = objTransactionOutput2.UnmarshalFrom(dataTransactionOutput)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objTransactionOutput, objTransactionOutput2) {
			panic(fmt.Sprintf("obj %T not equal", objTransactionOutput))
		}

	case 10:
		objTransactionWitness := &TransactionWitness{}
		_, err := objTransactionWitness.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataTransactionWitness, err := objTransactionWitness.Marshal()
		if err != nil {
			panic(err)
		}

		objTransactionWitness2 := &TransactionWitness{}
		_, err = objTransactionWitness2.UnmarshalFrom(dataTransactionWitness)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objTransactionWitness, objTransactionWitness2) {
			panic(fmt.Sprintf("obj %T not equal", objTransactionWitness))
		}

	case 11:
		objTransactions := &Transactions{}
		_, err := objTransactions.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataTransactions, err := objTransactions.Marshal()
		if err != nil {
			panic(err)
		}

		objTransactions2 := &Transactions{}
		_, err = objTransactions2.UnmarshalFrom(dataTransactions)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objTransactions, objTransactions2) {
			panic(fmt.Sprintf("obj %T not equal", objTransactions))
		}

	case 12:
		objTransferTransaction := &TransferTransaction{}
		_, err := objTransferTransaction.UnmarshalFrom(data[1:])
		if err != nil {
			return 0
		}
		dataTransferTransaction, err := objTransferTransaction.Marshal()
		if err != nil {
			panic(err)
		}

		objTransferTransaction2 := &TransferTransaction{}
		_, err = objTransferTransaction2.UnmarshalFrom(dataTransferTransaction)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(objTransferTransaction, objTransferTransaction2) {
			panic(fmt.Sprintf("obj %T not equal", objTransferTransaction))
		}
	}

	return 1
}

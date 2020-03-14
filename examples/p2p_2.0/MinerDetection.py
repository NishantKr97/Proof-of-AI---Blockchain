import pickle
import sklearn
import csv 
import pandas as pd
from sklearn import preprocessing

def readCSV():
	# csv file name 
	filename = "minerdets.csv"
	data = pd.read_csv (r'minerdets.csv')
	df = pd.DataFrame(data)
	# df[0] = ["User ID", "No. of Blocks Mined", "Faulty Transactions", "Time to Mine", "Age"]  # adding a row
	print (df)
	return df

def convertUserIDtoInt(X):
	le = preprocessing.LabelEncoder()
	le.fit(X['User ID'])
	list(le.classes_)
	# le.transform(Y)
	X['User ID'] = le.transform(X['User ID'])
	print(X['User ID'])
	return X

def converIntToUserID(X_test):
	return le.inverse_transform(X_test['User ID'])


def main():
	df = readCSV()
	X = df
	print(X['Time to Mine'])
	# Converting User ID to Int format
	le = preprocessing.LabelEncoder()
	le.fit(X['User ID'])
	list(le.classes_)
	X['User ID'] = le.transform(X['User ID'])


	print(X['User ID'])

	print(X['Age'])

	# Using the saved Model
	loaded_model = pickle.load(open('perceptron.sav', 'rb'))
	X_test = X

	# Saving the predicted value
	y_predicted = loaded_model.predict(X_test)
	print("Result : ", y_predicted)


	# Retrieving the User ID from the Int value
	Q = le.inverse_transform(X_test['User ID'])
	Final_Array = pd.DataFrame({'User ID': Q, 'Output' : y_predicted})
	print(Final_Array.head())

	# Selecting the Miner
	miner = ""
	for index, row in Final_Array.iterrows():
		if row['Output'] == 1:
			miner = row['User ID']
			#break 

	print(miner)	

	# Storing the Result in a file
	file1 = open("miner.txt","w") 
	file1.writelines(miner) 
	file1.close() 


if __name__ == '__main__':
	main()



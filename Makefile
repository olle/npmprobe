.PHONY: prep
prep:
	cat compromised.txt | sort | uniq > temp.txt
	cp temp.txt compromised.txt
	rm temp.txt


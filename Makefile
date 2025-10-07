##
## Normalize list of compromised packages by sorting and removing
## duplicate entries. Use this target before updating the repository.
##
.PHONY: prepare-list
prepare-list:
	cat compromised.txt | sort | uniq > temp.txt
	cp temp.txt compromised.txt
	rm temp.txt

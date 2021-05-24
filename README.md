Transform a mysqldump into a json file, the JSON file into a dot file, and the dot file into a ERD.

```bash
go install github.com/simon-engledew/sqljson/cmd/sqljsondump@v0.0.3
go install github.com/simon-engledew/sqljson/cmd/sqljsondot@v0.0.3

mysqldump database | sqljsondump | sqljsondot | dot -Tpng -o database.png && open database.png
```

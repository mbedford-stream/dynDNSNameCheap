# NameCheap Dynamic DNS record updater.  
  
A very simple script to replace an even more simple python script I wrote years ago to keep my dynmic DNS entry up to date.  
  
# config file  
When running the script, "-c filename.ext" is needed to provide relevant info and should be in the following format:  

```{  
    "updateParams": {  
        "host": "test",  
        "domain": "mgmt.network",  
        "password": "NameCheap provided password",  
        "log": true,  
        "debug": false  
    }  
}  
```
The fields should be self explanatory. "log" will create a updateLog.txt file with the result of any time the script is run and "debug" prints to the terminal as things happen so you can keep track of the variables being checked in case things aren't working as you expect.  
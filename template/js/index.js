$(document).ready(function(){
	$("#files").bind('change', handleFileSelect);
	
	$("#load_file_next").click(handleSubmit);

	$("#set_revise").click(function(){
		$.post("/config", { data_revise: $('#set_revise').is(':checked') } );
	});

	$("#set_result_dir").click(function(){
		$.post("/config", { data_result: this.value } );
	});
	$("#set_result_mem").click(function(){
		$.post("/config", { data_result: this.value } );
	});

	$("#set_max_num").change(function(){
		$.post("/config", { data_max_num: this.value } );
	});

});

function handleFileSelect(evt) {
	var files = evt.target.files;

	$("#file_list").html("");

	for (var i=0; i<files.length; i++) {
		var file = files[i];

		// 提示载入文件的信息
		filelist.add(file, i+1, "#file_list");

		/* Instantiate the File Reader object. */
		var reader = new FileReader();

		/* onLoad event is fired when the load completes. */
		reader.onload = function(event) {
			//document.getElementById('content').textContent = event.target.result; 
			
			var data = event.target.result.toString();

			if (file.type==="text/plain" || file.type==="application/vnd.ms-excel") {
				var lines = data.split(/\r\n/);
				var title = lines[0];
				console.log(title);

				//title.indexof("SDR") !=-1 || title.indexof("Multi-Track Export") != -1
				if  (contains(title, "SDR") || contains(title, "Multi-Track Export")) {
					return
				}
			}

			new Message("error", "格式不符的数据文件！").show("#load_file .message").autohide();
			$("#load_file_next").addClass("disabled");

		};

		/* The readAsText method will read the file's data as a text string. By default the string is decoded as 'UTF-8'. */
		reader.readAsText(file);
	}
	$("#load_file_next").removeClass("disabled");
}

function handleSubmit() {
	if ($("#load_file_next").hasClass("disabled")) {
		new Message("error", "请载入正确的数据文件！").show("#load_file .message").autohide();
		return;
	}

	$("#file-submit").click();
}

/*
*string:原始字符串
*substr:子字符串
*isIgnoreCase:忽略大小写
*/

function contains(string,substr,isIgnoreCase){
	if(isIgnoreCase)	{
		string=string.toLowerCase();
		substr=substr.toLowerCase();
	}
	var startChar=substr.substring(0,1);
	var strLen=substr.length;
	for(var j=0;j<string.length-strLen+1;j++)	{
		if(string.charAt(j)==startChar){
			if(string.substring(j,j+strLen)==substr){
				return true;
			}
		}
	}
	return false;
}
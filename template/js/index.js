$(document).ready(function(){
	$("#files").bind('change', handleFileSelect);
	
	$("#load_file_next").click(handleSubmit);

	$("#set_revise").click(function(){
		$('#set_revise').is(':checked') ? $("#data_revise").val("true") : $("#data_revise").val("false");

		console.log($("#data_revise").val());
	});
});

function handleFileSelect(evt) {
	var files = evt.target.files;

	$("#file_list").html("");

	/*
	if (files.length%2 != 0) {
		new Message("error", "同时需要1小时和10分钟数据文件！").show("#load_file .message").autohide();
		$("#load_file_next").addClass("disabled");
		return
	}
	*/

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

			/*
			var lines = data.split(/\r\n/);
			var r = lines[0].split(/\t/);
			var system = r[0];

			var info = {
				"SDR": true
			}

			if (typeof(info[system])=="undefined") {
				new Message("error", "格式不符的数据文件！").show("#load_file .message").autohide();
				$("#load_file_next").addClass("disabled");
				return;
			}
			*/

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
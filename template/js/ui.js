/*== Filelist ==*/
var filelist = {
	add: function(file, index, list) {
		var content = "";
		var alertId = "msg_"+index;

		$(list).append('<div id="'+alertId+'"></div>');

		content = content + '文件'+index+'：<strong>' + escape(file.name) + '</strong> (' +
				 (file.type || 'n/a') + ') - ' +
					file.size/1000 + ' Kb, 修改日期: ' +
					file.lastModifiedDate.toLocaleDateString();

		var type = "success";
		if (file.type!="text/plain") {
			type = "error";
		}
		new Message(type, content).show("#"+alertId);
	}
}

/*== Message ==*/
function Message (type, content, id) {
	if (typeof(id)=="undefined") {
		id = newId();
	}
	this.id = id;
	this.content = '<div id="'+this.id+'" class="alert alert-block alert-'+type+' fade in">'+content+'</div>';
}
Message.prototype.autohide = function(time) {
	if (typeof(time)=="undefined") {
		time = 3000;
	}
	setTimeout("$('#"+this.id+"').addClass('hidden')", time);
	return this;
}
Message.prototype.show = function(id) {
	$(id).html(this.content);
	return this;
}

function newId() {
	var guid = "";
	for (var i = 1; i <= 32; i++){
		var n = Math.floor(Math.random()*16.0).toString(16);
		guid +=   n;
		if((i==8)||(i==12)||(i==16)||(i==20))
			guid += "-";
	}
	return guid;
}
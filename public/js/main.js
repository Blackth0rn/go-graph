// Main.js - for the main things in life

var connection = new WebSocket('ws://localhost:8080/ws');
connection.binaryType="arraybuffer";

var output = document.getElementById('messages');
var start_state = document.getElementById('start_state');
var action = document.getElementById('action');
var end_state = document.getElementById('end_state');
var send = document.getElementById('send');
var msg_type = document.getElementById('msg_type');

var const_error			= 0;
var const_add_link		= 1;
var const_send_list		= 2;
var const_send_clear	= 3;

send.addEventListener('click', sendMessage, false);

function display_add_link(data)
{
	span = document.createElement("span");
	span.innerHTML = "Data: " + JSON.stringify(decoded_data);
	output.appendChild(span);
	output.appendChild(document.createElement("br"));
}

function clearInput()
{
	for( input in [start_state, action, end_state] )
	{
		input.value = '';
	}
}

function getInput(input_array)
{
	data = {};
	for( input in input_array )
	{
		data[input_array[input].id] = input_array[input].value;
	}
	return data;
}

function sendMessage()
{
	var data = [];
	var type = parseInt(msg_type.value,10);
	switch(type)
	{
		case const_add_link:
			data = sendLink();
			break;
		case const_send_list:
			data = sendList();
			break;
		case const_send_clear:
			data = sendClear();
			break;
	}
	// encode with msgpack here
	// make the packed data
	packed_data = msgpack.pack(data, false);
	// make a buffer big enough to hold the packed data (length * bytes per element) + 1 for the type
	buffer = new ArrayBuffer(packed_data.length * Uint8Array.BYTES_PER_ELEMENT + 1);
	//make a view for our buffer
	data_array = new Uint8Array(buffer);
	// copy our msg_pack into our buffer, offset of 1 to leave room for the msg_type
	data_array.set(packed_data, 1);
	data_array.set([type], 0);
	connection.send(data_array);
}

function sendLink()
{
	data = getInput([start_state, action, end_state]);
	clearInput();
	return data;
}

connection.onmessage = function(e) {
	data = Array.from(new Uint8Array(e.data));
	msg_type = data[0];
	packed_data = data.slice(1);
	decoded_data = msgpack.unpack(packed_data);
	switch(msg_type)
	{
		case const_error:
			display_error(decoded_data);
			break;
		case const_add_link:
			display_add_link(decoded_data);
			break;
		case const_send_list:
			display_send_list(decoded_data);
			break;
		case const_send_clear:
			display_send_clear(decoded_data);
			break;
	}
}


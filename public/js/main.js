// Main.js - for the main things in life

var connection = new WebSocket('ws://localhost:8080/ws');
connection.binaryType="arraybuffer";

var output = document.getElementById('messages');
var start_state = document.getElementById('start_state');
var action = document.getElementById('action');
var end_state = document.getElementById('end_state');
var send = document.getElementById('send');

send.addEventListener('click', sendMessage, false);

function displayMessage(data)
{
	span = document.createElement("span");
	// decode msgpack here
	decoded_data = msgpack.unpack(data)
	span.innerHTML = JSON.stringify(decoded_data);
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
	data = getInput([start_state, action, end_state]);
	clearInput();
	// encode with msgpack here
	data_array = Uint8Array.from(msgpack.pack(data, false));
	connection.send(data_array);
}

connection.onmessage = function(e) {
	data = Array.from(new Uint8Array(e.data));
	displayMessage(data);
}


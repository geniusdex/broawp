function webSocketEndpoint()
{
    var protocol = ((window.location.protocol === "https:") ? "wss:" : "ws:");
    var basepath = window.location.pathname.substr(0, window.location.pathname.lastIndexOf("/"));
    return protocol + "//" + window.location.host + basepath + "/ws";
}

let g_webSocket = new WebSocket(webSocketEndpoint());

g_webSocket.addEventListener('error', function(event)
{
    console.log(event);
});

g_webSocket.addEventListener('close', function(event)
{
    console.log(event);
});

let messageHandlers = {};

g_webSocket.addEventListener('message', function(event)
{
    var messages = event.data.split('\n');
    for (var i = 0; i < messages.length; i++)
    {
        if (messages[i].length > 0)
        {
            var msg = JSON.parse(messages[i]);
            // console.log(msg);
            if ('type' in msg && 'data' in msg && msg.type in messageHandlers)
            {
                try
                {
                    messageHandlers[msg.type](msg.data);
                }
                catch (error)
                {
                    console.error(error, msg);
                }
            }
            else
            {
                // console.log('Ignoring unexpected message', msg);
            }
        }
    }
});

export function newHandler(type, handler)
{
    messageHandlers[type] = handler;
}

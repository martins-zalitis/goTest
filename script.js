// const request = require('/usr/local/node_modules/request');

module.exports = {
    /*
     User test script START
     */


    'RTC test' : function(client) {var demoAppUrl = 'https://dab6fab6.ngrok.io/localrec.html';
//var roomName = '0712TESTS_';

client
    .url(demoAppUrl)
    .waitForElementVisible('body', 1000)
    .saveScreenshot('./screenshots/sel.png')
    .pause(10000)
    .saveScreenshot('./screenshots/selEnd.png')

    ;},

    /*
     User test script END
     */

    after : function(client) {
        client
            .pause(1000)
            .end();
    },
};

'use strict';

const express = require('express');
var expressJWT = require('express-jwt');
var jwt = require('jsonwebtoken');
var bearerToken = require('express-bearer-token');
var cors = require('cors');
var session = require('express-session');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var util = require('util');

require('./configSetting')
const invoke = require('./invoke-transaction')
const query = require('./query')
const helper = require('./helper')
var logger = helper.getLogger('server');

// Constants
const PORT = 8080;
const HOST = '0.0.0.0';

function getErrorMessage(field) {
	var response = {
		success: false,
		message: field + ' field is missing or Invalid in the request'
	};
	return response;
}

// App
const app = express();

///////////////////////////////////////////////////////////////////////////////
//////////////////////////////// SET CONFIGURATONS ////////////////////////////
///////////////////////////////////////////////////////////////////////////////
app.options('*', cors());
app.use(cors());
//support parsing of application/json type post data
app.use(bodyParser.json());
//support parsing of application/x-www-form-urlencoded post data
app.use(bodyParser.urlencoded({
	extended: false
}));
// set secret variable
app.set('secret', 'thisismysecret');
app.use(expressJWT({
	secret: 'thisismysecret'
}).unless({
	path: ['/signin']
}));
app.use(bearerToken());
app.use(function(req, res, next) {
	logger.debug(' ------>>>>>> new request for %s',req.originalUrl);
	if (req.originalUrl.indexOf('/signin') >= 0) {
		return next();
	}

	var token = req.token;
	jwt.verify(token, app.get('secret'), function(err, decoded) {
		if (err) {
			res.send({
				success: false,
				message: 'Failed to authenticate token. Make sure to include the ' +
					'token returned from /users call in the authorization header ' +
					' as a Bearer token'
			});
			return;
		} else {
			// add the decoded user name and org name to the request object
			// for the downstream code to use
			req.username = decoded.username;
			req.orgname = decoded.orgName;
			logger.debug(util.format('Decoded from JWT token: username - %s, orgname - %s', decoded.username, decoded.orgName));
			return next();
		}
	});
});

const orgFredrick = 'fredrick'
const orgRegulator = 'regulator'
const orgAlice = 'alice'
const orgBob = 'bob'
const fishermans = [orgFredrick]
const restauranteurs = [orgAlice, orgBob]
const regulators = [orgRegulator]
const allOrgs = [orgFredrick, orgAlice, orgBob]

app.get('/', (req, res) => {
  res.send('Home\n');
});

// signing in as an org
// not requiring password for now
app.post('/signin', async function(req, res) {
  try {
    var orgName = req.body.orgName;
    logger.debug('End point : /login');
    logger.debug('Org name  : ' + orgName);
    if (!orgName) {
      res.json(getErrorMessage('\'orgName\''));
      return;
    }

    if (allOrgs.includes(orgName)) {
      const username = 'user1' // just use user1 as default for now

      var token = jwt.sign({
        exp: Math.floor(Date.now() / 1000) + 36000,
        username,
        orgName: orgName
      }, app.get('secret'));
      let response = await helper.getRegisteredUser(username, orgName, true);
      logger.debug('-- returned from get the username %s for organization %s',username,orgName);
      if (response && typeof response !== 'string') {
        logger.debug('Successfully GET the username %s for organization %s',username,orgName);
        response.token = token;
        res.json(response);
      } else {
        logger.debug('Failed to get the username %s for organization %s with::%s',username,orgName,response);
        res.json({success: false, message: response});
      }
    } else if (regulators.includes(orgName)) {
      var token = jwt.sign({
        exp: Math.floor(Date.now() / 1000) + 36000,
        username: 'regulator',
        orgName: 'regulator',
      }, app.get('secret'));
      res.json({success: true, token});
    } else {
      res.json({
        success: false,
        message: "orgName doesn't exist."
      })
      return
    }
    
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
});

// anyone signed in can call this endpoint
app.get('/salmons/:id', async (req, res) => {
  logger.debug('username :' + req.username);
  logger.debug('orgname:' + req.orgname);
  try {
    const channelName = 'transfers'
    const chaincodeName = 'salmon'
    const fcn = "querySalmon"
    const args = [req.params.id]
    logger.debug('salmon_id:' + req.params.id);

    let message = await query.queryChaincode([], channelName, chaincodeName, args, fcn, req.username, req.orgname);
    res.send(message);
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
})

// only regulator can call this endpoint
app.get('/salmons', async (req, res) => {
  logger.debug('orgname:' + req.orgname);
  if (!regulators.includes(req.orgname)) {
    res.status(401).json({
      message: "You're not authorized for this information"
    })
  }

  try {
    const channelName = 'transfers'
    const chaincodeName = 'salmon'
    const fcn = "queryAllSalmon"
    const args = []

    let message = await query.queryChaincode([], channelName, chaincodeName, args, fcn, 'user1', 'fredrick'); // use fisherman user to retrieve the data because regulator doesn't have an user
    res.send(message);
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
})

// only fisherman can call this
app.post('/salmons', async (req, res) => {
  try {
    logger.debug('username :' + req.username);
    logger.debug('orgname:' + req.orgname);

    if (!fishermans.includes(req.orgname)) {
      res.json({
        success: false,
        message: "You're not authorized for this action"
      })
      return
    }
    
    const {
      id,
      vessel,
      datetime,
      location,
    } = req.body

    if (!id) {
      res.json(getErrorMessage('\'id\''));
      return;
    }
    if (!vessel) {
      res.json(getErrorMessage('\'vessel\''));
      return;
    }

    const channelName = 'transfers'
    const chaincodeName = 'salmon'
    const fcn = "recordSalmon"
    const args = [
      id,
      vessel,
      datetime,
      location,
      req.orgname, // holder
    ]
    logger.debug('args: ', args)

    const message = await invoke.invokeChaincode(['peer0.fredrick.coderschool.vn'], channelName, chaincodeName, fcn, args, req.username, req.orgname);
    res.send(message);
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
});

// only restauranteur can call this
app.post('/salmons/:salmon_id/ownerships/claim', async (req, res) => {
  logger.debug('username :' + req.username);
  logger.debug('orgname:' + req.orgname);

  try {
    if (!restauranteurs.includes(req.orgname)) {
      res.json({
        success: false,
        message: "You're not authorized for this action"
      })
      return
    }

    const id = req.params.salmon_id

    if (!id) {
      res.json(getErrorMessage('\'id\''));
      return;
    }

    let agreement, changeHolderTrxId
    {
      const channelName = `fredrick-${ req.orgname }`
      const chaincodeName = 'agreement'
      const fcn = "queryAgreement"
      const args = [
        "1",
      ]

      const _agreement = await query.queryChaincode([], channelName, chaincodeName, args, fcn, req.username, req.orgname);
      agreement = JSON.parse(_agreement)
    }

    {
      const channelName = 'transfers'
      const chaincodeName = 'salmon'
      const fcn = "changeSalmonHolder"
      const args = [
        id,
        req.orgname, // new holder
      ]

      changeHolderTrxId = await invoke.invokeChaincode(['peer0.fredrick.coderschool.vn'], channelName, chaincodeName, fcn, args, req.username, req.orgname);
    }

    res.send({
      agreement,
      changeHolderTrxId,
    });
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
});

// only regulator can call this
app.post('/orgs/:orgname/users/:username/revoke', async (req, res) => {
  logger.debug('orgname:' + req.orgname);
  if (!regulators.includes(req.orgname)) {
    res.status(401).json({
      message: "You're not authorized for this information"
    })
  }

  try {
    const message = await helper.revokeUser(req.params.username, req.params.orgname)
    res.send(message);
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
})

// only regulator can call this
app.post('/orgs/:orgname/users/:username/enroll', async (req, res) => {
  logger.debug('orgname:' + req.orgname);
  if (!regulators.includes(req.orgname)) {
    res.status(401).json({
      message: "You're not authorized for this information"
    })
  }

  try {
    const {username, orgname} = req.params
    const response = await helper.getRegisteredUser(username, orgname, true)
    logger.debug('-- returned from registering the username %s for organization %s',username,orgname);
    if (response && typeof response !== 'string') {
      logger.debug('Successfully registered the username %s for organization %s',username,orgname);
      res.json(response);
    } else {
      logger.debug('Failed to register the username %s for organization %s with::%s',username,orgname,response);
      res.json({success: false, message: response});
    }
  } catch (err) {
    console.error(err)
    res.status(500).send({error: err.message})
  }
})

app.listen(PORT, HOST);
console.log(`Running on http://${HOST}:${PORT}`);
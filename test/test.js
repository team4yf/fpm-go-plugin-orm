const { fpmc, DBQuery } = require('fpmc-jssdk');
fpmc.init({ appkey: '123123', masterKey: '123123', endpoint: 'http://localhost:9090/api', v: '0.0.1' });

(async ()=>{
    try {
        const query = new DBQuery('fake')
        const data = await query.select('NAME').condition(`name = 'ff'`).page(1,10).sort('name-').first()
        
        
        console.log(data)    
    } catch (error) {
        console.error(error)
    }
    
})()
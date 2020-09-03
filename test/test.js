const { fpmc, DBQuery } = require('fpmc-jssdk');
const { DBObject } = require('fpmc-jssdk/lib/bo/impl/DBObject');
fpmc.init({ appkey: '123123', masterKey: '123123', endpoint: 'http://localhost:9090/api', v: '0.0.1' });

(async ()=>{
    try {
        // const query = new DBQuery('fake')
        // const data = await query.select('NAME').condition(`name = 'ff'`).page(1,10).sort('name-').first()
        // console.log(data)   
        const obj = new DBObject('fake', {id: 110}) 
        // const rsp = await obj.save({
        //     name: 'c'
        // })
        
        // console.log(rsp)
        console.log(await obj.remove(112))
        
    } catch (error) {
        console.error(error)
    }
    
})()
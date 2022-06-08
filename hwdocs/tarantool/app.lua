box.cfg {
    listen = 3301;
    memtx_memory = 128 * 1024 * 1024; -- 128Mb
    memtx_min_tuple_size = 16;
    memtx_max_tuple_size = 128 * 1024 * 1024; -- 128Mb
    vinyl_memory = 128 * 1024 * 1024; -- 128Mb
    vinyl_cache = 128 * 1024 * 1024; -- 128Mb
    vinyl_max_tuple_size = 128 * 1024 * 1024; -- 128Mb
    vinyl_write_threads = 2;
    wal_mode = "none";
    wal_max_size = 256 * 1024 * 1024;
    checkpoint_interval = 60 * 60; -- one hour
    checkpoint_count = 6;
    force_recovery = true;

     -- 1 – SYSERROR
     -- 2 – ERROR
     -- 3 – CRITICAL
     -- 4 – WARNING
     -- 5 – INFO
     -- 6 – VERBOSE
     -- 7 – DEBUG
     log_level = 4;
     too_long_threshold = 0.5;
 }

box.cfg({memtx_memory = box.cfg.memtx_memory + 712 * 2^20})
box.schema.user.grant('guest','read,write,execute','universe')

local function bootstrap()

    if not box.space.mysql_profile then
        t = box.schema.space.create('mysql_profile')
        t:create_index('userid',
            {type = 'tree', parts = {1, 'unsigned'}, unique = true, if_not_exists = true})
    end

    if not box.space.mysql_interest then
        t2 = box.schema.space.create('mysql_interest')
        t2:create_index('primary',
            {type = 'tree', parts = {1, 'unsigned'}, if_not_exists = true})
        t2:create_index('userid',
            {type = 'tree', parts = {2, 'unsigned'}, unique = false, if_not_exists = true})
    end

    if not box.space.mysql_sex then
        t2 = box.schema.space.create('mysql_sex')
        t2:create_index('primary',
            {type = 'tree', parts = {1, 'unsigned'}, if_not_exists = true})
    end

    function load_profile(user_id)
        local data = box.space.mysql_profile.index.userid:select({user_id})
        if #data == 0 then
            return {}
        end

        local interests = {}

        local interestsTable = box.space.mysql_interest.index.userid:select({user_id})
        if #interestsTable ~= 0 then
            for i, v in ipairs(interestsTable) do
                interests[i] = {}
                interests[i].id = v[1]
                interests[i].name = v[3]
            end
        end
        
        local sexData = box.space.mysql_sex.index.primary:select({data[1][6]})
        local profile = {}
        profile.id = data[1][2]
        profile.name = data[1][3]
        profile.surname = data[1][4]
        profile.birth_date = {timeStamp(data[1][5]), 0}
        profile.sex = {}
        profile.sex.id = sexData[1][1]
        profile.sex.name = sexData[1][2]
        profile.city = data[1][7]
        
        local user = {}
        user.id = user_id
        user.interests = interests
        user.profile = profile
        
        return user
    end
    
    function timeStamp(dateStringArg)
        local inYear, inMonth, inDay = string.match(dateStringArg, '^(%d%d%d%d)-(%d%d)-(%d%d)$')
        return os.time({year=inYear, month=inMonth, day=inDay, hour=0, min=0, sec=0, isdst=false})
    end
end

bootstrap()
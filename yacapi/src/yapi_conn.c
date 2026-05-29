#include "yapi_inc.h"
#include "stdlib.h"
#include "string.h"

YapiResult yapiAllocConnect(YapiEnv* env, YapiConnect** hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiConnect* conn;
    if (yapiAllocMem("Connection", 1, sizeof(YapiConnect), (void**)&conn, &error) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }
    if (yapiCliAllocHandle(YAPI_HANDLE_DBC, env->envHandler, &conn->connHandler, &error) != YAPI_SUCCESS) {
        yapiFreeMem(conn);
        return YAPI_ERROR;
    }

    *hConn = conn;
    return YAPI_SUCCESS;
}

YapiResult yapiConnect2(YapiConnect* hConn, const char* url, int16_t urlLength, const char* user, int16_t userLength,
                       const char* password, int16_t passwordLength)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    if (*url == '\0') {
        uint8_t credType = CRED_EXT;
        if (yapiSetConnAttr(hConn, YAPI_ATTR_CREDT, &credType, sizeof(uint8_t)) != YAPI_SUCCESS) {
            yapiCliFreeHandle(YAPI_HANDLE_DBC, hConn->connHandler, &error);
            yapiFreeMem(hConn);
            return YAPI_ERROR;
        }
    }
    if (yapiCliConnect(hConn->connHandler, url, urlLength, user, userLength, password, passwordLength, &error) !=
        YAPI_SUCCESS) {
        yapiCliFreeHandle(YAPI_HANDLE_DBC, hConn->connHandler, &error);
        yapiFreeMem(hConn);
        return YAPI_ERROR;
    }
    return YAPI_SUCCESS;
}

YapiResult yapiConnect(YapiEnv* env, const char* url, int16_t urlLength, const char* user, int16_t userLength,
                       const char* password, int16_t passwordLength, YapiConnect** hConn)
{
    if (yapiAllocConnect(env, hConn) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }

    return yapiConnect2(*hConn, url, urlLength, user, userLength, password, passwordLength);
}

YapiResult yapiDisconnect(YapiConnect* hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliDisconnect(hConn->connHandler, &error);
}

YapiResult yapiReleaseConn(YapiConnect* hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    YAPI_CALL(yapiCliFreeHandle(YAPI_HANDLE_DBC, hConn->connHandler, &error));
    yapiFreeMem(hConn);

    return YAPI_SUCCESS;
}

YapiResult yapiCancel(YapiConnect* hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliCancel(hConn->connHandler, &error);
}

YapiResult yapiCommit(YapiConnect* hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliCommit(hConn->connHandler, &error);
}

YapiResult yapiRollback(YapiConnect* hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliRollback(hConn->connHandler, &error);
}

YapiResult yapiSetConnAttr(YapiConnect* hConn, YapiConnAttr attr, void* value, int32_t length)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliSetConnAttr(hConn->connHandler, attr, value, length, &error);
}

YapiResult yapiGetConnAttr(YapiConnect* hConn, YapiConnAttr attr, void* value, int32_t bufLength, int32_t* stringLength)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliGetConnAttr(hConn->connHandler, attr, value, bufLength, stringLength, &error);
}

YapiResult yapiPing(YapiConnect* hConn, int32_t timeout)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliPing(hConn->connHandler, timeout, &error);
}

YapiResult yapiParseSqlParams(YapiEnv* hEnv, YapiPointer* paramList, const char* sql, int32_t sqlLength)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliParseSqlParams(hEnv->envHandler, paramList, sql, sqlLength, &error);
}
YapiResult yapiGetParamListCount(YapiPointer hParamList, uint32_t* count)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliGetParamListCount(hParamList, count, &error);
}

YapiResult yapiGetParamName(YapiPointer hParamList, uint16_t index, char* name, int32_t nameBufLen, int32_t* nameLen)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliGetParamName(hParamList, index, name, nameBufLen, nameLen, &error);
}

YapiResult yapiFreeParamList(YapiPointer hParamList)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliFreeParamList(hParamList, &error);
}

YapiResult yapiGetSqlParamCount(const char* sql, int32_t sqlLength, uint16_t* paramCount)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliGetSqlParamCount(sql, sqlLength, paramCount, &error);
}

void yapiGetLastError(YapiErrorInfo* info)
{
    YapiErrorMsg error;

    yapiInitError(&error);
    yapiGetErrorInfo(&error, info);
}

YapiResult yapiLobDescAlloc(YapiConnect* hConn, YapiType type, void** desc)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliLobDescAlloc(hConn->connHandler, type, desc, &error);
}

YapiResult yapiLobDescFree(void* desc, YapiType type)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliLobDescFree(desc, type, &error);
}

YapiResult yapiLobGetChunkSize(YapiConnect* hConn, YapiLobLocator* locator, uint16_t* chunkSize)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliLobGetChunkSize(hConn->connHandler, locator, chunkSize, &error);
}

YapiResult yapiLobGetLength(YapiConnect* hConn, YapiLobLocator* locator, uint64_t* length)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliLobGetLength(hConn->connHandler, locator, length, &error);
}

YapiResult yapiLobRead(YapiConnect* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliLobRead(hConn->connHandler, loc, bytes, buf, bufLen, &error);
}

YapiResult yapiLobWrite(YapiConnect* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliLobWrite(hConn->connHandler, loc, bytes, buf, bufLen, &error);
}

YapiResult yapiLobCreateTemporary(YapiConnect* hConn, YapiLobLocator* loc)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliLobCreateTemporary(hConn->connHandler, loc, &error);
}

YapiResult yapiLobFreeTemporary(YapiConnect* hConn, YapiLobLocator* loc)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliLobFreeTemporary(hConn->connHandler, loc, &error);
}

YapiResult yapiAllocConnectionPool(YapiEnv* env, YapiConnectPool** hConnPool)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiConnectPool* connPool;
    if (yapiAllocMem("ConnectionPool", 1, sizeof(YapiConnectPool), (void**)&connPool, &error) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }
    if (yapiCliAllocHandle(YAPI_HANDLE_CONN_POOL, env->envHandler, &connPool->connPoolHandler, &error) != YAPI_SUCCESS) {
        yapiFreeMem(connPool);
        return YAPI_ERROR;
    }

    connPool->env = env;
    *hConnPool = connPool;
    return YAPI_SUCCESS;
}

YapiResult yapiReleaseConnectionPool(YapiConnectPool* hConnPool)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiResult ret = yapiCliFreeHandle(YAPI_HANDLE_CONN_POOL, hConnPool->connPoolHandler, &error);
    yapiFreeMem(hConnPool);
    return ret;
}

YapiResult yapiConnectionPoolCreate(YapiConnectPool* hConnPool, const char* url, int16_t urlLength,
                                    uint32_t min, uint32_t max, uint32_t increment, const char* user, int16_t userLength,
                                    const char* password, int16_t passwordLength, uint32_t mode)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliConnectionPoolCreate(hConnPool->connPoolHandler, url, urlLength, min, max, increment,
                                    user, userLength, password, passwordLength, mode, &error);
}

YapiResult yapiConnectionGet(YapiConnectPool* hConnPool, YapiConnect** hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiConnect* conn;
    if (yapiAllocMem("Connection", 1, sizeof(YapiConnect), (void**)&conn, &error) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }
    if (yapiCliConnectionGet(hConnPool->connPoolHandler, &conn->connHandler, &error) != YAPI_SUCCESS) {
        yapiFreeMem(conn);
        return YAPI_ERROR;
    }

    *hConn = conn;
    return YAPI_SUCCESS;
}

YapiResult yapiConnectionGiveBack(YapiConnect* hConn)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiResult ret = yapiCliConnectionGiveBack(hConn->connHandler, &error);
    yapiFreeMem(hConn);
    return ret;
}

YapiResult yapiConnectionPoolDestroy(YapiConnectPool* hConnPool, uint32_t mode)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliConnectionPoolDestroy(hConnPool->connPoolHandler, mode, &error);
}

YapiResult yapiDescAlloc2(YapiEnv* hEnv, void** desc, YapiDescType type)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliDescAlloc2(hEnv->envHandler, desc, type, &error);
}

YapiResult yapiDescFree2(YapiEnv* hEnv, void* desc, YapiDescType type)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliDescFree2(hEnv->envHandler, desc, type, &error);
}

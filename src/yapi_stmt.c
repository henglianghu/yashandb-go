#include "yapi_inc.h"
#include "stdlib.h"

YapiResult yapiStmtCreate(YapiConnect* hConn, YapiStmt** hStmt) 
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiStmt* stmt;
    if (yapiAllocMem("Statment", 1, sizeof(YapiStmt), (void**)&stmt, &error) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }
    if (yapiCliAllocHandle(YAPI_HANDLE_STMT, hConn->connHandler, &stmt->stmtHandler, &error) != YAPI_SUCCESS) {
        yapiFreeMem(stmt);
        return YAPI_ERROR;
    }

    *hStmt = stmt;
    return YAPI_SUCCESS;
}

YapiResult yapiCursorStmtCreate(YapiConnect* hConn, YapiStmt** hStmt, YacHandle *handle)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiStmt* stmt;
    if (yapiAllocMem("Statment", 1, sizeof(YapiStmt), (void**)&stmt, &error) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }
    
    stmt->stmtHandler = *handle;
    *hStmt = stmt;
    return YAPI_SUCCESS;
}




YapiResult yapiStmtPrepare(YapiStmt* hStmt, const char* sql, int32_t sqlLength) 
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliPrepare(hStmt->stmtHandler, sql, sqlLength, &error);
}

YapiResult yapiPrepare(YapiConnect* hConn, const char* sql, int32_t sqlLength, YapiStmt** hStmt)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    YapiStmt* stmt;
    if (yapiAllocMem("Statment", 1, sizeof(YapiStmt), (void**)&stmt, &error) != YAPI_SUCCESS) {
        return YAPI_ERROR;
    }
    if (yapiCliAllocHandle(YAPI_HANDLE_STMT, hConn->connHandler, &stmt->stmtHandler, &error) != YAPI_SUCCESS) {
        yapiFreeMem(stmt);
        return YAPI_ERROR;
    }
    if (yapiCliPrepare(stmt->stmtHandler, sql, sqlLength, &error) != YAPI_SUCCESS) {
        yapiCliFreeHandle(YAPI_HANDLE_STMT, stmt->stmtHandler, &error);
        yapiFreeMem(stmt);
        return YAPI_ERROR;
    }
    *hStmt = stmt;
    return YAPI_SUCCESS;
}

YapiResult yapiReleaseStmt(YapiStmt* hStmt)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    yapiCliFreeHandle(YAPI_HANDLE_STMT, hStmt->stmtHandler, &error);
    free(hStmt);
    return YAPI_SUCCESS;
}


YapiResult yapiReleaseHandle(YacHandle* handle)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    yapiCliFreeHandle(YAPI_HANDLE_STMT, handle, &error);
    return YAPI_SUCCESS;
}


YapiResult yapiReleaseCursorStmt(YapiStmt* hStmt)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    free(hStmt);
    return YAPI_SUCCESS;
}

YapiResult yapiExecute(YapiStmt* hStmt)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliExecute(hStmt->stmtHandler, &error);
}

YapiResult yapiFetch(YapiStmt* hStmt, uint32_t* rows)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliFetch(hStmt->stmtHandler, rows, &error);
}

YapiResult yapiDirectExecute(YapiStmt* hStmt, const char* sql, int32_t sqlLength)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliDirectExecute(hStmt->stmtHandler, sql, sqlLength, &error);
}

YapiResult yapiDescribeCol2(YapiStmt* hStmt, uint16_t id, YapiColumnDesc* desc)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliDescribeCol2(hStmt->stmtHandler, id, desc, &error);
}

YapiResult yapiBindColumn(YapiStmt* hStmt, uint16_t id, YapiType type, YapiPointer value, int32_t bufLen,
                        int32_t* indicator)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliBindColumn(hStmt->stmtHandler, id, type, value, bufLen, indicator, &error);
}

YapiResult yapiBindParameter(YapiStmt* hStmt, uint16_t id, YapiParamDirection direction, YapiType bindType,
                           YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliBindParameter(hStmt->stmtHandler, id, direction, bindType, value, bindSize, bufLength, indicator,
                                &error);
}

YapiResult yapiBindParameterByName(YapiStmt* hStmt, char* name, YapiParamDirection direction, YapiType bindType,
                                   YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliBindParameterByName(hStmt->stmtHandler, name, direction, bindType, value, bindSize, bufLength,
                                      indicator, &error);
}

YapiResult yapiNumResultCols(YapiStmt* hStmt, int16_t* count)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliNumResultCols(hStmt->stmtHandler, count, &error);
}

YapiResult yapiSetStmtAttr(YapiStmt* hStmt, YapiStmtAttr attr, void* value, int32_t length)
{
    YapiErrorMsg error;
    yapiInitError(&error);
    return yapiCliSetStmtAttr(hStmt->stmtHandler, attr, value, length, &error);
}

YapiResult yapiGetStmtAttr(YapiStmt* hStmt, YapiStmtAttr attr, void* value, int32_t bufLength, int32_t* stringLength)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliGetStmtAttr(hStmt->stmtHandler, attr, value, bufLength, stringLength, &error);
}

YapiResult yapiColAttribute(YapiStmt* hStmt, uint16_t id, YapiColAttr attr, void* value, int32_t bufLen,
                               int32_t* stringLength)
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliColAttribute(hStmt->stmtHandler, id, attr, value, bufLen, stringLength, &error);
}

YapiResult yapiNumParams(YapiStmt* hStmt, int16_t* count) 
{
    YapiErrorMsg error;
    yapiInitError(&error);

    return yapiCliNumParams(hStmt->stmtHandler, count, &error);
}
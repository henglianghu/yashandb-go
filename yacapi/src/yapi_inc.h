#ifndef YAPI_INC_H
#define YAPI_INC_H

#include "yacapi.h"

#ifdef __cplusplus
extern "C" {
#endif

#define YAPI_CALL(yapiFunc)      \
    do {                         \
        YapiResult r = yapiFunc; \
        if (r == YAPI_ERROR) {   \
            return r;            \
        }                        \
    } while (0)

typedef enum EnYacResult { YAC_SUCCESS = 0, YAC_SUCCESS_WITH_INFO = 1, YAC_ERROR = -1 } YacResult;

typedef YacResult (*yapiFuncAllocHandle)(YapiHandleType type, YacHandle input, YacHandle* output);
typedef YacResult (*yapiFuncFreeHandler)(YapiHandleType type, YacHandle handle);
typedef char* (*yapiFuncGetVersion)();
typedef void (*yapiFuncGetLastError)(int32_t* errCode, char** message, char** sqlState, YapiTextPos* pos);
typedef YacResult (*yapiFuncSetEnvAttr)(YacHandle hEnv, YapiEnvAttr attr, void* value, int32_t length);
typedef YacResult (*yapiFuncGetEnvAttr)(YacHandle hEnv, YapiEnvAttr attr, void* value, int32_t bufLength,
                                        int32_t* stringLength);

typedef YacResult (*yapiFuncConnect)(YacHandle hConn, const char* url, int16_t urlLength, const char* user,
                                     int16_t userLength, const char* password, int16_t passwordLength);
typedef void (*yapiFuncDisconnect)(YacHandle hConn);
typedef YacResult (*yapiFuncCancel)(YacHandle hConn);
typedef YacResult (*yapiFuncDirectExecute)(YacHandle hStmt, const char* sql, int32_t sqlLength);
typedef YacResult (*yapiFuncPrepare)(YacHandle hStmt, const char* sql, int32_t sqlLength);
typedef YacResult (*yapiFuncExecute)(YacHandle hStmt);
typedef YacResult (*yapiFuncFetch)(YacHandle hStmt, uint32_t* rows);
typedef YacResult (*yapiFuncCommit)(YacHandle hConn);
typedef YacResult (*yapiFuncRollback)(YacHandle hConn);
typedef YacResult (*yapiFuncPing)(YacHandle hConn, int32_t timeout);

typedef YacResult (*yapiFuncParseSqlParams)(YacHandle hEnv, YacHandle* hParamList, const char* sql, int32_t sqlLength);
typedef YacResult (*yapiFuncGetParamListCount)(YacHandle hParamList, uint32_t* count);
typedef YacResult (*yapiFuncGetParamName)(YacHandle hParamList, uint16_t index, char* name, int32_t nameBufLen, int32_t* nameLen);
typedef YacResult (*yapiFuncFreeParamList)(YacHandle hParamList);
typedef YacResult (*yapiFuncGetSqlParamCount)(const char* sql, int32_t sqlLength, uint16_t* paramCount);

typedef YacResult (*yapiFuncSetConnAttr)(YacHandle hConn, YapiConnAttr attr, void* value, int32_t length);
typedef YacResult (*yapiFuncGetConnAttr)(YacHandle hConn, YapiConnAttr attr, void* value, int32_t bufLength,
                                         int32_t* stringLength);
typedef YacResult (*yapiFuncSetStmtAttr)(YacHandle hStmt, YapiStmtAttr attr, void* value, int32_t length);
typedef YacResult (*yapiFuncGetStmtAttr)(YacHandle hStmt, YapiStmtAttr attr, void* value, int32_t bufLength,
                                         int32_t* stringLength);

typedef YacResult (*yapiFuncDescribeCol2)(YacHandle hStmt, uint16_t id, YapiColumnDesc* desc);
typedef YacResult (*yapiFuncBindColumn)(YacHandle hStmt, uint16_t id, YapiType type, void* value, int32_t bufLen,
                                        int32_t* indicator);
typedef YacResult (*yapiFuncBindParameter)(YacHandle hStmt, uint16_t id, YapiParamDirection direction,
                                           YapiType bindType, void* value, int32_t bindSize, int32_t bufLength,
                                           int32_t* indicator);
typedef YacResult (*yapiFuncBindParameterByName)(YacHandle hStmt, char* name, YapiParamDirection direction,
                                                 YapiType bindType, void* value, int32_t bindSize, int32_t bufLength,
                                                 int32_t* indicator);
typedef YacResult (*yapiFuncNumResultCols)(YacHandle hStmt, int16_t* count);
typedef YacResult (*yapiFuncColAttribute)(YacHandle hStmt, uint16_t id, YapiColAttr attr, void* value, int32_t bufLen,
                                          int32_t* stringLength);
typedef YacResult (*yapiFuncNumParams)(YacHandle hStmt, int16_t* count);

typedef YapiDate (*yapiFuncNow)();
typedef void (*yapiFuncNumberFromInt32)(YapiNumber* n, int32_t v);
typedef YacResult (*yapiFuncText2Timestamp)(char* text, char* format, YapiDate* stamp, int16_t* bias);
typedef YacResult (*yapiFuncText2YMInterval)(char* str, YapiYMInterval* interval);
typedef YacResult (*yapiFuncText2DSInterval)(char* str, YapiDSInterval* interval);
typedef YacResult (*yapiFuncText2ShortTime)(char* str, char* format, YapiShortTime* shortTime);
typedef YacResult (*yapiFuncGetDateStruct)(YapiDate date, YapiDateStruct* ds);

// multi insert API in "batch" mode
typedef YacResult (*yapiFuncBatchInsertPrepare)(YacHandle hStmt, char* tableName);
typedef YacResult (*yapiFuncBatchInsertExecute)(YacHandle hStmt);

// lob API
typedef YacResult (*yapiFuncLobDescAlloc)(YacHandle hConn, YapiType type, void** desc);
typedef YacResult (*yapiFuncLobDescFree)(void* desc, YapiType type);
typedef YacResult (*yapiFuncLobGetChunkSize)(YacHandle hConn, YapiLobLocator* locator, uint16_t* chunkSize);
typedef YacResult (*yapiFuncLobGetLength)(YacHandle hConn, YapiLobLocator* locator, uint64_t* length);
typedef YacResult (*yapiFuncLobRead)(YacHandle hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf,
                                     uint64_t bufLen);
typedef YacResult (*yapiFuncLobWrite)(YacHandle hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf,
                                      uint64_t bufLen);
typedef YacResult (*yapiFuncLobCreateTemporary)(YacHandle hConn, YapiLobLocator* loc);
typedef YacResult (*yapiFuncLobFreeTemporary)(YacHandle hConn, YapiLobLocator* loc);

// vector API
typedef YacResult (*yapiFuncDescAlloc2)(YacHandle hEnv, void** desc, YapiDescType type);
typedef YacResult (*yapiFuncDescFree2)(YacHandle hEnv, void* desc, YapiDescType type);
typedef YacResult (*yapiFuncVectorFromText)(YapiVector* vector, YapiVectorFormat format, uint16_t dim, char* text, uint32_t textlen, uint32_t mode);
typedef YacResult (*yapiFuncVectorFromArray)(YapiVector* vector, YapiVectorFormat format, uint16_t dim, uint8_t* array, uint32_t arrayLen, uint32_t mode);
typedef YacResult (*yapiFuncVectorToText)(YapiVector* vector, char* text, uint32_t* textlen, uint32_t mode);
typedef YacResult (*yapiFuncVectorToArray)(YapiVector* vector, YapiVectorFormat format, uint16_t* dim, uint8_t* array, uint32_t* arrayLen, uint32_t mode);
typedef YacResult (*yapiFuncVectorGetFormat)(YapiVector* vector, YapiVectorFormat* format);
typedef YacResult (*yapiFuncVectorGetDimension)(YapiVector* vector, uint16_t* dim);

// dataType API
typedef YacResult (*yapiFuncDateGetDate)(const YapiDate date, int16_t* year, uint8_t* month, uint8_t* day);
typedef YacResult (*yapiFuncShortTimeGetShortTime)(const YapiShortTime time, uint8_t* hour, uint8_t* minute,
                                                   uint8_t* second, uint32_t* fraction);
typedef YacResult (*yapiFuncTimestampGetTimestamp)(const YapiTimestamp timestamp, int16_t* year, uint8_t* month,
                                                   uint8_t* day, uint8_t* hour, uint8_t* minute, uint8_t* second,
                                                   uint32_t* fraction);
typedef YacResult (*yapiFuncYMIntervalGetYearMonth)(const YapiYMInterval ymInterval, int32_t* year, int32_t* month);
typedef YacResult (*yapiFuncDSIntervalGetDaySecond)(const YapiDSInterval dsInterval, int32_t* day, int32_t* hour,
                                                    int32_t* minute, int32_t* second, int32_t* fraction);

typedef YacResult (*yapiFuncDateSetDate)(YapiDate* date, int16_t year, uint8_t month, uint8_t day);
typedef YacResult (*yapiFuncShortTimeSetShortTime)(YapiShortTime* time, uint8_t hour, uint8_t minute, uint8_t second,
                                                   uint32_t fraction);
typedef YacResult (*yapiFuncTimestampSetTimestamp)(YapiTimestamp* timestamp, int16_t year, uint8_t month, uint8_t day,
                                                   uint8_t hour, uint8_t minute, uint8_t second, uint32_t fraction);
typedef YacResult (*yapiFuncYMIntervalSetYearMonth)(YapiYMInterval* ymInterval, int32_t year, int32_t month);
typedef YacResult (*yapiFuncDSIntervalSetDaySecond)(YapiDSInterval* dsInterval, int32_t day, int32_t hour,
                                                    int32_t minute, int32_t second, int32_t fraction);

typedef YacResult (*yapiFuncDateTimeGetTimeZoneOffset)(YacHandle hEnv, YapiTimestamp timestamp, int8_t* hr, int8_t* mm);

typedef YacResult (*yapiFuncDSIntervalFromText)(YacHandle hEnv, YapiDSInterval* dsInterval, const char* str,
                                                uint32_t strLen);

typedef YacResult (*yapiFuncYMIntervalFromText)(YacHandle hEnv, YapiYMInterval* ymInterval, const char* str,
                                                uint32_t strLen);

typedef YacResult (*yapiFuncNumberRound)(YapiNumber* n, int32_t precision, int32_t scale);

typedef YacResult (*yapiFuncNumberFromText)(const char* str, uint32_t strLength, const char* fmt, uint32_t fmtLength,
                                            const char* nlsParam, uint32_t nlsParamLength, YapiNumber* number);

typedef YacResult (*yapiFuncNumberToText)(const YapiNumber* number, const char* fmt, uint32_t fmtLength,
                                          const char* nlsParam, uint32_t nlsParamLength, char* str, int32_t bufLength,
                                          int32_t* length);

typedef YacResult (*yapiFuncNumberFromReal)(const YapiPointer rnum, uint32_t length, YapiNumber* number);
typedef YacResult (*yapiFuncNumberToReal)(const YapiNumber* number, uint32_t length, YapiPointer rsl);

typedef YacResult (*yapiFuncPdbgStart)(YacHandle stmt, uint64_t objId, uint16_t subId);

typedef YacResult (*yapiFuncPdbgCheckVersion)(YapiStmt* hStmt, uint64_t objId, uint16_t subId, uint32_t version);

typedef YacResult (*yapiFuncPdbgAbort)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgContinue)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgStepInto)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgStepOut)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgStepNext)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgShowSource)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgDeleteAllBreakpoints)(YacHandle stmt);

typedef YacResult (*yapiFuncPdbgAddBreakpoint)(YacHandle stmt, uint64_t objId, uint16_t subId, uint32_t lineNo,
                                               uint32_t* bpId);

typedef YacResult (*yapiFuncPdbgDeleteBreakpoint)(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo);
typedef YacResult (*yapiFuncPdbgGetBreakpointsCount)(YacHandle hStmt, uint32_t* bpCount);
typedef YacResult (*yapiFuncPdbgGetAllVars)(YacHandle stmt, uint32_t* varCount);
typedef YacResult (*yapiFuncPdbgGetAllFrames)(YacHandle hStmt, uint32_t* frameCount);

typedef YacResult (*yapiFuncPdbgGetRunningAttrs)(YacHandle hStmt, YapiDebugRunningAttr attr, void* value,
                                                 int32_t bufLen, int32_t* stringLength);
typedef YacResult (*yapiFuncPdbgGetFrameAttrs)(YacHandle hStmt, uint32_t id, YapiDebugFrameAttr attr, void* value,
                                               int32_t bufLen, int32_t* stringLength);
typedef YacResult (*yapiFuncPdbgGetVarAttrs)(YacHandle hStmt, uint32_t id, YapiDebugVarAttr attr, void* value,
                                             int32_t bufLen, int32_t* stringLength);
typedef YacResult (*yapiFuncPdbgGetVarValue)(YacHandle hStmt, uint32_t id, uint32_t valueType, void* value,
                                             int32_t bufLen, int32_t* indicator);
typedef YacResult (*yapiFuncPdbgGetBreakpointAttrs)(YacHandle hStmt, uint32_t id, YapiDebugBpAttr attr, void* value,
                                                    int32_t bufLen, int32_t* stringLength);

typedef YacResult (*yapiFuncConnectionPoolCreate)(YacHandle hConnPool, const char* url, int16_t urlLength,
                                                  uint32_t min, uint32_t max, uint32_t increment, const char* user, int16_t userLength,
                                                  const char* password, int16_t passwordLength, uint32_t mode);
typedef YacResult (*yapiFuncConnectionGet)(YacHandle hConnPool, YacHandle hConn);
typedef YacResult (*yapiFuncConnectionGiveBack)(YacHandle hConn);
typedef YacResult (*yapiFuncConnectionPoolDestroy)(YacHandle hConnPool, uint32_t mode);

typedef struct StYapiSymbols {
    yapiFuncAllocHandle fnAllocHandle;
    yapiFuncFreeHandler fnHandleFree;

    yapiFuncConnect    fnConnect;
    yapiFuncDisconnect fnDisconnect;

    yapiFuncGetVersion   fnGetVersion;
    yapiFuncGetLastError fnGetLastError;

    yapiFuncCancel        fnCancel;
    yapiFuncDirectExecute fnDirectExecute;
    yapiFuncPrepare       fnPrepare;
    yapiFuncExecute       fnExecute;
    yapiFuncFetch         fnFetch;
    yapiFuncCommit        fnCommit;
    yapiFuncRollback      fnRollback;
    yapiFuncPing          fnPing;

    yapiFuncParseSqlParams     fnParseSqlParams;
    yapiFuncGetParamListCount  fnGetParamListCount;
    yapiFuncGetParamName       fnGetParamName;
    yapiFuncFreeParamList      fnFreeParamList;
    yapiFuncGetSqlParamCount   fnGetSqlParamCount;

    yapiFuncSetEnvAttr  fnSetEnvAttr;
    yapiFuncGetEnvAttr  fnGetEnvAttr;
    yapiFuncSetConnAttr fnSetConnAttr;
    yapiFuncGetConnAttr fnGetConnAttr;
    yapiFuncSetStmtAttr fnSetStmtAttr;
    yapiFuncGetStmtAttr fnGetStmtAttr;

    yapiFuncDescribeCol2        fnDescribeCol2;
    yapiFuncBindColumn          fnBindColumn;
    yapiFuncBindParameter       fnBindParameter;
    yapiFuncBindParameterByName fnBindParameterByName;
    yapiFuncNumResultCols       fnNumResultCols;
    yapiFuncColAttribute        fnColAttribute;
    yapiFuncNumParams           fnNumParams;

    yapiFuncNow             fnNow;
    yapiFuncNumberFromInt32 fnNumberFromInt32;
    yapiFuncText2Timestamp  fnText2Timestamp;
    yapiFuncText2YMInterval fnText2YMInterval;
    yapiFuncText2DSInterval fnText2DSInterval;
    yapiFuncText2ShortTime  fnText2ShortTime;
    yapiFuncGetDateStruct   fnGetDateStruct;

    yapiFuncBatchInsertPrepare fnBatchInsertPrepare;
    yapiFuncBatchInsertExecute fnBatchInsertExecute;

    yapiFuncLobDescAlloc       fnLobDescAlloc;
    yapiFuncLobDescFree        fnLobDescFree;
    yapiFuncLobGetChunkSize    fnLobGetChunkSize;
    yapiFuncLobGetLength       fnLobGetLength;
    yapiFuncLobRead            fnLobRead;
    yapiFuncLobWrite           fnLobWrite;
    yapiFuncLobCreateTemporary fnLobCreateTemporary;
    yapiFuncLobFreeTemporary   fnLobFreeTemporary;

    yapiFuncDateGetDate            fnDateGetDate;
    yapiFuncShortTimeGetShortTime  fnShortTimeGetShortTime;
    yapiFuncTimestampGetTimestamp  fnTimestampGetTimestamp;
    yapiFuncYMIntervalGetYearMonth fnYMIntervalGetYearMonth;
    yapiFuncDSIntervalGetDaySecond fnDSIntervalGetDaySecond;

    yapiFuncDateSetDate            fnDateSetDate;
    yapiFuncShortTimeSetShortTime  fnShortTimeSetShortTime;
    yapiFuncTimestampSetTimestamp  fnTimestampSetTimestamp;
    yapiFuncYMIntervalSetYearMonth fnYMIntervalSetYearMonth;
    yapiFuncDSIntervalSetDaySecond fnDSIntervalSetDaySecond;
    yapiFuncDSIntervalFromText     fnDSIntervalFromText;
    yapiFuncYMIntervalFromText     fnYMIntervalFromText;

    yapiFuncDateTimeGetTimeZoneOffset fnDateTimeGetTimeZoneOffset;

    yapiFuncNumberRound    fnNumberRound;
    yapiFuncNumberFromText fnNumberFromText;
    yapiFuncNumberToText   fnNumberToText;
    yapiFuncNumberFromReal fnNumberFromReal;
    yapiFuncNumberToReal   fnNumberToReal;

    yapiFuncPdbgStart        fnPdbgStart;
    yapiFuncPdbgCheckVersion fnPdbgCheckVersion;
    yapiFuncPdbgAbort        fnPdbgAbort;
    yapiFuncPdbgContinue     fnPdbgContinue;
    yapiFuncPdbgStepInto     fnPdbgStepInto;
    yapiFuncPdbgStepOut      fnPdbgStepOut;
    yapiFuncPdbgStepNext     fnPdbgStepNext;

    yapiFuncPdbgDeleteAllBreakpoints fnPdbgDeleteAllBreakpoints;
    yapiFuncPdbgAddBreakpoint        fnPdbgAddBreakpoint;
    yapiFuncPdbgDeleteBreakpoint     fnPdbgDeleteBreakpoint;
    yapiFuncPdbgGetBreakpointsCount  fnPdbgGetBreakpointsCount;
    yapiFuncPdbgGetAllVars           fnPdbgPdbgGetAllVars;
    yapiFuncPdbgGetAllFrames         fnPdbgGetAllFrames;

    yapiFuncPdbgGetRunningAttrs    fnPdbgGetRunningAttrs;
    yapiFuncPdbgGetFrameAttrs      fnPdbgGetFrameAttrs;
    yapiFuncPdbgGetVarValue        fnPdbgGetVarValue;
    yapiFuncPdbgGetVarAttrs        fnPdbgGetVarAttrs;
    yapiFuncPdbgGetBreakpointAttrs fnPdbgGetBreakpointAttrs;

    yapiFuncConnectionPoolCreate   fnConnectionPoolCreate;
    yapiFuncConnectionGet          fnConnectionGet;
    yapiFuncConnectionGiveBack      fnConnectionGiveBack;
    yapiFuncConnectionPoolDestroy  fnConnectionPoolDestroy;

    yapiFuncDescAlloc2          fnDescAlloc2;
    yapiFuncDescFree2           fnDescFree2;
    yapiFuncVectorFromText      fnVectorFromText;
    yapiFuncVectorFromArray     fnVectorFromArray;
    yapiFuncVectorToText        fnVectorToText;
    yapiFuncVectorToArray       fnVectorToArray;
    yapiFuncVectorGetFormat     fnVectorGetFormat;
    yapiFuncVectorGetDimension  fnVectorGetDimension;

} YapiSymbols;

#define T2S_BUFFER_SIZE 8192

typedef struct StYapiErrorBuffer {
    int32_t     code;
    uint32_t    messageLen;
    char        message[T2S_BUFFER_SIZE];
    char        sqlState[YAPI_MAX_SQLSTAT_LEN];
    YapiTextPos pos;
} YapiErrorBuffer;

typedef struct StYapiErrorMsg {
    YapiErrorBuffer* buf;
} YapiErrorMsg;

typedef enum {
    YAPI_ERR_NO_ERR = 20000,
    YAPI_ERR_LOAD_SYMBOL,
    YAPI_ERR_ALLOC_MEM,
} yapiErrorNum;

typedef struct StYapiEnv {
    uint32_t  version;
    YacHandle envHandler;
} YapiEnv;

typedef struct StYapiConnect {
    YapiEnv*  env;
    YacHandle connHandler;
} YapiConnect;

typedef struct StYapiStmt {
    YapiConnect* conn;
    YacHandle    stmtHandler;
} YapiStmt;

typedef struct StYapiConnectPool {
    YapiEnv*  env;
    YacHandle connPoolHandler;
} YapiConnectPool;

YapiResult yapiOpenDynamicLib(char* libName, YapiPointer* handler, YapiErrorMsg* error);
void       yapiSetError(YapiErrorMsg* error, yapiErrorNum errorNum, const char* format, ...);

YapiResult yapiCliAllocHandle(YapiHandleType type, YacHandle input, YacHandle* output, YapiErrorMsg* error);
YapiResult yapiCliFreeHandle(YapiHandleType type, YacHandle handle, YapiErrorMsg* error);
YapiResult yapiCliGetVersion(char** version, YapiErrorMsg* error);
YapiResult yapiCliGetLastError(YapiErrorMsg* error);

YapiResult yapiCliSetEnvAttr(YapiEnv* hEnv, YapiEnvAttr attr, void* value, int32_t length, YapiErrorMsg* error);
YapiResult yapiCliGetEnvAttr(YacHandle hEnv, YapiEnvAttr attr, void* value, int32_t bufLength, int32_t* stringLength,
                             YapiErrorMsg* error);

YapiResult yapiCliConnect(YacHandle hConn, const char* url, int16_t urlLength, const char* user, int16_t userLength,
                          const char* password, int16_t passwordLength, YapiErrorMsg* error);
YapiResult yapiCliDisconnect(YacHandle hConn, YapiErrorMsg* error);
YapiResult yapiCliCommit(YacHandle hConn, YapiErrorMsg* error);
YapiResult yapiCliRollback(YacHandle hConn, YapiErrorMsg* error);
YapiResult yapiCliSetConnAttr(YacHandle hConn, YapiConnAttr attr, void* value, int32_t length, YapiErrorMsg* error);
YapiResult yapiCliGetConnAttr(YacHandle hConn, YapiConnAttr attr, void* value, int32_t bufLength, int32_t* stringLength,
                              YapiErrorMsg* error);
YapiResult yapiCliCancel(YacHandle hConn, YapiErrorMsg* error);
YapiResult yapiCliPing(YacHandle hConn, int32_t timeout, YapiErrorMsg* error);

YapiResult yapiCliParseSqlParams(YacHandle hEnv, YacHandle* paramList, const char* sql, int32_t sqlLength, YapiErrorMsg* error);
YapiResult yapiCliGetParamListCount(YacHandle hParamList, uint32_t* count, YapiErrorMsg* error);
YapiResult yapiCliGetParamName(YacHandle hParamList, uint16_t index, char* name, int32_t nameBufLen, int32_t* nameLen, YapiErrorMsg* error);
YapiResult yapiCliFreeParamList(YacHandle hParamList, YapiErrorMsg* error);
YapiResult yapiCliGetSqlParamCount(const char* sql, int32_t sqlLength, uint16_t* paramCount, YapiErrorMsg* error);

YapiResult yapiCliDirectExecute(YacHandle hStmt, const char* sql, int32_t sqlLength, YapiErrorMsg* error);
YapiResult yapiCliPrepare(YacHandle hStmt, const char* sql, int32_t sqlLength, YapiErrorMsg* error);
YapiResult yapiCliExecute(YacHandle hStmt, YapiErrorMsg* error);
YapiResult yapiCliSetStmtAttr(YacHandle hStmt, YapiStmtAttr attr, void* value, int32_t length, YapiErrorMsg* error);
YapiResult yapiCliGetStmtAttr(YacHandle hStmt, YapiStmtAttr attr, void* value, int32_t bufLength, int32_t* stringLength,
                              YapiErrorMsg* error);
YapiResult yapiCliFetch(YacHandle hStmt, uint32_t* rows, YapiErrorMsg* error);
YapiResult yapiCliDescribeCol2(YacHandle hStmt, uint16_t id, YapiColumnDesc* desc, YapiErrorMsg* error);
YapiResult yapiCliBindColumn(YacHandle hStmt, uint16_t id, YapiType type, YapiPointer value, int32_t bufLen,
                             int32_t* indicator, YapiErrorMsg* error);
YapiResult yapiCliBindParameter(YacHandle hStmt, uint16_t id, YapiParamDirection direction, YapiType bindType,
                                YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator,
                                YapiErrorMsg* error);
YapiResult yapiCliBindParameterByName(YacHandle hStmt, char* name, YapiParamDirection direction, YapiType bindType,
                                      YapiPointer value, int32_t bindSize, int32_t bufLength, int32_t* indicator,
                                      YapiErrorMsg* error);
YapiResult yapiCliNumResultCols(YacHandle hStmt, int16_t* count, YapiErrorMsg* error);
YapiResult yapiCliColAttribute(YacHandle hStmt, uint16_t id, YapiColAttr attr, void* value, int32_t bufLen,
                               int32_t* stringLength, YapiErrorMsg* error);
YapiResult yapiCliNumParams(YacHandle hStmt, int16_t* count, YapiErrorMsg* error);

YapiResult yapiCliGetDateStruct(YapiDate date, YapiDateStruct* ds, YapiErrorMsg* error);

YapiResult yapiCliLobDescAlloc(YacHandle* hConn, YapiType type, void** desc, YapiErrorMsg* error);
YapiResult yapiCliLobDescFree(void* desc, YapiType type, YapiErrorMsg* error);
YapiResult yapiCliLobGetChunkSize(YacHandle* hConn, YapiLobLocator* locator, uint16_t* chunkSize, YapiErrorMsg* error);
YapiResult yapiCliLobGetLength(YacHandle* hConn, YapiLobLocator* locator, uint64_t* length, YapiErrorMsg* error);
YapiResult yapiCliLobRead(YacHandle* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen,
                          YapiErrorMsg* error);
YapiResult yapiCliLobWrite(YacHandle* hConn, YapiLobLocator* loc, uint64_t* bytes, uint8_t* buf, uint64_t bufLen,
                           YapiErrorMsg* error);
YapiResult yapiCliLobCreateTemporary(YacHandle* hConn, YapiLobLocator* loc, YapiErrorMsg* error);
YapiResult yapiCliLobFreeTemporary(YacHandle* hConn, YapiLobLocator* loc, YapiErrorMsg* error);

YapiResult yapiCliDateGetDate(const YapiDate date, int16_t* year, uint8_t* month, uint8_t* day, YapiErrorMsg* error);
YapiResult yapiCliShortTimeGetShortTime(const YapiShortTime time, uint8_t* hour, uint8_t* minute, uint8_t* second,
                                        uint32_t* fraction, YapiErrorMsg* error);
YapiResult yapiCliTimestampGetTimestamp(const YapiTimestamp timestamp, int16_t* year, uint8_t* month, uint8_t* day,
                                        uint8_t* hour, uint8_t* minute, uint8_t* second, uint32_t* fraction,
                                        YapiErrorMsg* error);
YapiResult yapiCliYMIntervalGetYearMonth(const YapiYMInterval ymInterval, int32_t* year, int32_t* month,
                                         YapiErrorMsg* error);
YapiResult yapiCliDSIntervalGetDaySecond(const YapiDSInterval dsInterval, int32_t* day, int32_t* hour, int32_t* minute,
                                         int32_t* second, int32_t* fraction, YapiErrorMsg* error);

YapiResult yapiCliDateSetDate(YapiDate* date, int16_t year, uint8_t month, uint8_t day, YapiErrorMsg* error);
YapiResult yapiCliShortTimeSetShortTime(YapiShortTime* time, uint8_t hour, uint8_t minute, uint8_t second,
                                        uint32_t fraction, YapiErrorMsg* error);
YapiResult yapiCliTimestampSetTimestamp(YapiTimestamp* timestamp, int16_t year, uint8_t month, uint8_t day,
                                        uint8_t hour, uint8_t minute, uint8_t second, uint32_t fraction,
                                        YapiErrorMsg* error);
YapiResult yapiCliDateTimeGetTimeZoneOffset(YacHandle hEnv, YapiTimestamp timestamp, int8_t* hr, int8_t* mm,
                                            YapiErrorMsg* error);

YapiResult yapiCliYMIntervalSetYearMonth(YapiYMInterval* ymInterval, int32_t year, int32_t month, YapiErrorMsg* error);
YapiResult yapiCliDSIntervalSetDaySecond(YapiDSInterval* dsInterval, int32_t day, int32_t hour, int32_t minute,
                                         int32_t second, int32_t fraction, YapiErrorMsg* error);

YapiResult yapiCliDSIntervalFromText(YacHandle hEnv, YapiDSInterval* dsInterval, const char* str, uint32_t strLen,
                                     YapiErrorMsg* error);

YapiResult yapiCliYMIntervalFromText(YacHandle hEnv, YapiYMInterval* ymInterval, const char* str, uint32_t strLen,
                                     YapiErrorMsg* error);

YapiResult yapiCliNumberRound(YapiNumber* n, int32_t precision, int32_t scale, YapiErrorMsg* error);

YapiResult yapiCliNumberFromText(const char* str, uint32_t strLength, const char* fmt, uint32_t fmtLength,
                                 const char* nlsParam, uint32_t nlsParamLength, YapiNumber* number,
                                 YapiErrorMsg* error);

YapiResult yapiCliNumberToText(const YapiNumber* number, const char* fmt, uint32_t fmtLength, const char* nlsParam,
                               uint32_t nlsParamLength, char* str, int32_t bufLength, int32_t* length,
                               YapiErrorMsg* error);

YapiResult yapiCliNumberFromReal(const YapiPointer rnum, uint32_t length, YapiNumber* number, YapiErrorMsg* error);

YapiResult yapiCliNumberToReal(const YapiNumber* number, uint32_t length, YapiPointer rsl, YapiErrorMsg* error);

void yapiInitError(YapiErrorMsg* error);
void yapiGetErrorInfo(YapiErrorMsg* error, YapiErrorInfo* info);
void yapiGetCliError(YapiErrorMsg* error);

YapiResult yapiAllocMem(const char* name, size_t numMembers, size_t memberSize, void** ptr, YapiErrorMsg* error);
void       yapiFreeMem(void* ptr);

YapiResult yapiCiPdbgStart(YacHandle hStmt, uint64_t objId, uint16_t subId, YapiErrorMsg* error);
YapiResult yapiCiPdbgCheckVersion(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t version,
                                  YapiErrorMsg* error);
YapiResult yapiCiPdbgAbort(YacHandle hStmt, YapiErrorMsg* error);
YapiResult yapiCiPdbgContinue(YacHandle hStmt, YapiErrorMsg* error);
YapiResult yapiCiPdbgStepInto(YacHandle hStmt, YapiErrorMsg* error);
YapiResult yapiCiPdbgStepOut(YacHandle hStmt, YapiErrorMsg* error);
YapiResult yapiCiPdbgStepNext(YacHandle hStmt, YapiErrorMsg* error);

YapiResult yapiCiPdbgDeleteAllBreakpoints(YacHandle hStmt, YapiErrorMsg* error);
YapiResult yapiCiPdbgAddBreakpoint(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo, uint32_t* bpId,
                                   YapiErrorMsg* error);
YapiResult yapiCiPdbgDeleteBreakpoint(YacHandle hStmt, uint64_t objId, uint16_t subId, uint32_t lineNo,
                                      YapiErrorMsg* error);
YapiResult yapiCiPdbgGetBreakpointsCount(YacHandle hStmt, uint32_t* bpCount, YapiErrorMsg* error);
YapiResult yapiCiPdbgGetAllVars(YacHandle hStmt, uint32_t* varCount, YapiErrorMsg* error);
YapiResult yapiCiPdbgGetAllFrames(YacHandle hStmt, uint32_t* frameCount, YapiErrorMsg* error);

YapiResult yapiCiPdbgGetRunningAttrs(YacHandle hStmt, YapiDebugRunningAttr attr, void* value, int32_t bufLen,
                                     int32_t* stringLength, YapiErrorMsg* error);
YapiResult yapiCiPdbgGetFrameAttrs(YacHandle hStmt, uint32_t id, YapiDebugFrameAttr attr, void* value, int32_t bufLen,
                                   int32_t* stringLength, YapiErrorMsg* error);
YapiResult yapiCiPdbgGetVarValue(YacHandle hStmt, uint32_t id, YapiDebugVarAttr attr, void* value, int32_t bufLen,
                                 int32_t* stringLength, YapiErrorMsg* error);
YapiResult yapiCiPdbgGetVarAttrs(YacHandle hStmt, uint32_t id, uint32_t valueType, void* value, int32_t bufLen,
                                 int32_t* indicator, YapiErrorMsg* error);
YapiResult yapiCiPdbgGetBreakpointAttrs(YacHandle hStmt, uint32_t id, YapiDebugBpAttr attr, void* value, int32_t bufLen,
                                        int32_t* stringLength, YapiErrorMsg* error);

YapiResult yapiCliConnectionPoolCreate(YacHandle hConnPool, const char* url, int16_t urlLength,
                                       uint32_t min, uint32_t max, uint32_t increment, const char* user, int16_t userLength,
                                       const char* password, int16_t passwordLength, uint32_t mode, YapiErrorMsg* error);
YapiResult yapiCliConnectionGet(YacHandle hConnPool, YacHandle* hConn, YapiErrorMsg* error);
YapiResult yapiCliConnectionGiveBack(YacHandle hConn, YapiErrorMsg* error);
YapiResult yapiCliConnectionPoolDestroy(YacHandle hConnPool, uint32_t mode, YapiErrorMsg* error);

YapiResult yapiCliDescAlloc2(YacHandle hEnv, void** desc, YapiDescType type, YapiErrorMsg* error);
YapiResult yapiCliDescFree2(YacHandle hEnv, void* desc, YapiDescType type, YapiErrorMsg* error);
YapiResult yapiCliVectorFromText(YapiVector* vector, YapiVectorFormat format, uint16_t dim, char* text, uint32_t textlen, uint32_t mode, YapiErrorMsg* error);
YapiResult yapiCliVectorFromArray(YapiVector* vector, YapiVectorFormat format, uint16_t dim, uint8_t* array, uint32_t arrayLen, uint32_t mode, YapiErrorMsg* error);
YapiResult yapiCliVectorToText(YapiVector* vector, char* text, uint32_t* textlen, uint32_t mode, YapiErrorMsg* error);
YapiResult yapiCliVectorToArray(YapiVector* vector, YapiVectorFormat format, uint16_t* dim, uint8_t* array, uint32_t* arrayLen, uint32_t mode, YapiErrorMsg* error);
YapiResult yapiCliVectorGetFormat(YapiVector* vector, YapiVectorFormat* format, YapiErrorMsg* error);
YapiResult yapiCliVectorGetDimension(YapiVector* vector, uint16_t* dim, YapiErrorMsg* error);

#ifdef __cplusplus
}
#endif

#endif

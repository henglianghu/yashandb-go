#ifdef WIN32
#include <windows.h>
#else
#include <unistd.h>
#endif

#include "yacapi.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

#define YAPI_TEST_CALL(yapiFunc)  \
    do {                          \
        YapiResult r = yapiFunc;  \
        if (r == YAPI_ERROR) {    \
            return r;             \
        }                         \
    } while (0)

typedef struct {
    YapiEnv* hEnv;
    YapiConnect* hDbc;
    YapiStmt* hStmt;
} YapiHandles;

YapiHandles gTestHandles;

char* url = "127.0.0.1:1688";
char* user = "sys";
char* pwd = "Cod-2022";

YapiResult testSingle() 
{
    YAPI_TEST_CALL(yapiAllocEnv(&gTestHandles.hEnv));
    YAPI_TEST_CALL(yapiConnect(gTestHandles.hEnv, url, YAPI_NULL_TERM_STR, user, YAPI_NULL_TERM_STR, pwd,
                               YAPI_NULL_TERM_STR, &gTestHandles.hDbc));

    YAPI_TEST_CALL(yapiPrepare(gTestHandles.hDbc, "select 1 from dual", YAPI_NULL_TERM_STR, &gTestHandles.hStmt));
    YAPI_TEST_CALL(yapiExecute(gTestHandles.hStmt));
    int32_t intOut = 0;
    YAPI_TEST_CALL(yapiBindColumn(gTestHandles.hStmt, 0, YAPI_TYPE_INTEGER, &intOut, 0, NULL));
    uint32_t rows;
    YAPI_TEST_CALL(yapiFetch(gTestHandles.hStmt, &rows));
    if (intOut != 1) {
        return YAPI_ERROR;
    }

    YAPI_TEST_CALL(yapiReleaseStmt(gTestHandles.hStmt));
    YAPI_TEST_CALL(yapiReleaseConn(gTestHandles.hDbc));
    YAPI_TEST_CALL(yapiReleaseEnv(gTestHandles.hEnv));

    return YAPI_SUCCESS;
}


YapiResult testDirect() 
{
    YAPI_TEST_CALL(yapiAllocEnv(&gTestHandles.hEnv));
    YAPI_TEST_CALL(yapiConnect(gTestHandles.hEnv, url, YAPI_NULL_TERM_STR, user, YAPI_NULL_TERM_STR, pwd,
                               YAPI_NULL_TERM_STR, &gTestHandles.hDbc));
    YAPI_TEST_CALL(yapiStmtCreate(gTestHandles.hDbc, &gTestHandles.hStmt));

    // YAPI_TEST_CALL(yapiPrepare(gTestHandles.hDbc, "select 1 from dual", YAPI_NULL_TERM_STR, &gTestHandles.hStmt));
    // YAPI_TEST_CALL(yapiExecute(gTestHandles.hStmt));
    YAPI_TEST_CALL(yapiDirectExecute(gTestHandles.hStmt, "select 1 from dual", YAPI_NULL_TERM_STR));
    int32_t intOut = 0;
    YAPI_TEST_CALL(yapiBindColumn(gTestHandles.hStmt, 0, YAPI_TYPE_INTEGER, &intOut, 0, NULL));
    uint32_t rows;
    YAPI_TEST_CALL(yapiFetch(gTestHandles.hStmt, &rows));
    if (intOut != 1) {
        return YAPI_ERROR;
    }

    YAPI_TEST_CALL(yapiReleaseStmt(gTestHandles.hStmt));
    YAPI_TEST_CALL(yapiReleaseConn(gTestHandles.hDbc));
    YAPI_TEST_CALL(yapiReleaseEnv(gTestHandles.hEnv));

    return YAPI_SUCCESS;
}

int main(int argc, char** argv) 
{
    YAPI_TEST_CALL(testSingle());
    printf("test 1 success!\n");
    YAPI_TEST_CALL(testDirect());
    printf("test 2 success!\n");
    return YAPI_SUCCESS;
}
package lock

import (
	"math/rand"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
)

var rs *redsync.Redsync

func InitPool(pool redis.Pool) {
	rs = redsync.New(pool)
}

// Func: how to use lock
// func GetWeeklySkillPool() func(*gin.Context) {
// 	return func(c *gin.Context) {
//   	------ TRY TO LOCK --------------
// 		mutex, err := lock.Lock("getskillpool")
//	  ------ CHECK IF LOCK FAILED -----------
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, common.ErrorResponse(http.StatusInternalServerError, "cannot lock the progress"))
// 			return
// 		} else {
//	 	-------- UNLOCK -------------
// 		defer lock.Unlock(mutex)
// 		}
//
// 		time.Sleep(5 * time.Second)
// 		c.JSON(http.StatusOK, common.SuccessResponse(config.CachedWeeklySkillPool))
// 	}
// }
//

// Lock, func: we lock the mutex 8 seconds (timeout), 32 tries, each try is random between 50 - 250ms apart, so user must wait random from 1.6s - 8s
func Lock(mutexId string) (*redsync.Mutex, error) {
	mutex := rs.NewMutex(mutexId)
	return simpleLock(mutex)
}

// LockTimeout, func: we lock the mutex n (SECONDs), 32 tries, each try is random between 50 - 250ms apart
func LockTimeout(mutexId string, timeout int) (*redsync.Mutex, error) {
	timeoutOption := redsync.WithExpiry(time.Duration(timeout) * time.Second)
	mutex := rs.NewMutex(mutexId, timeoutOption)
	return simpleLock(mutex)
}

// Lock, func: we lock the mutex 8 seconds (timeout), n tries, each try is random between 50 - 250ms apart
func LockCustomRetry(mutexId string, retryCount int) (*redsync.Mutex, error) {
	retryOption := redsync.WithTries(retryCount)
	mutex := rs.NewMutex(mutexId, retryOption)
	return simpleLock(mutex)
}

// LockRetryDurationCustom, func: we lock the mutex 8 seconds (timeout), n tries, each try is eachTryDuration apart
func LockRetryDurationCustom(mutexId string, retryCount int, eachTryDuration time.Duration) (*redsync.Mutex, error) {
	retryOption := redsync.WithTries(retryCount)
	durationOption := redsync.WithRetryDelay(eachTryDuration)
	mutex := rs.NewMutex(mutexId, retryOption, durationOption)
	return simpleLock(mutex)
}

// LockRetryDurationTimeout, func: we lock the mutex <timeout> seconds, n tries, each try is eachTryDuration apart
func LockRetryDurationTimeout(mutexId string, retryCount int, eachTryDuration time.Duration, timeout time.Duration, opts ...redsync.Option) (*redsync.Mutex, error) {
	retryOption := redsync.WithTries(retryCount)
	durationOption := redsync.WithRetryDelay(eachTryDuration)
	timeoutOption := redsync.WithExpiry(timeout)
	mutex := rs.NewMutex(mutexId, retryOption, durationOption, timeoutOption)
	return simpleLock(mutex)
}

// LockRetryRandom, func: we lock the mutex 8 seconds (timeout), n tries and each try is random between min - max (ms) apart
func LockRetryRandom(mutexId string, retryCount int, minDuration int, maxDuration int) (*redsync.Mutex, error) {
	retryOption := redsync.WithTries(retryCount)
	duration := rand.Intn(maxDuration-minDuration) + minDuration

	durationOption := redsync.WithRetryDelay(time.Duration(duration) * time.Millisecond)
	mutex := rs.NewMutex(mutexId, retryOption, durationOption)
	return simpleLock(mutex)
}

func LockCustom(mutexId string, options ...redsync.Option) (*redsync.Mutex, error) {
	mutex := rs.NewMutex(mutexId, options...)
	return simpleLock(mutex)
}

func simpleLock(mutex *redsync.Mutex) (*redsync.Mutex, error) {
	if err := mutex.Lock(); err != nil {
		return nil, err
	}

	return mutex, nil
}

func Unlock(mutex *redsync.Mutex) (bool, error) {
	return mutex.Unlock()
}

// func Ab(tries int) time.Duration {
// 	return 1 * time.Second
// }

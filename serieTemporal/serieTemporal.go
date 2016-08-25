package serieTemporal

import (
    "time"
    "log"
)

type SerieTemporal struct {
    msecondsPerPoint int
    maxPoints int
    maxMSeconds int64
    data []Slot
}

type Slot struct {
    MSeconds int64
    Value float32
}

func NewSerieTemporal(msecondsPerPoint int, maxPoints int) *SerieTemporal {
    s := &SerieTemporal{msecondsPerPoint, maxPoints, int64(msecondsPerPoint*maxPoints), make([]Slot, maxPoints)}

    return s
}

func (s *SerieTemporal) UpdateNow(value float32) {
    s.Update(value,time.Now().Unix()*1000)
}

func (s *SerieTemporal) Update(value float32, mseconds int64) {
    now := time.Now().Unix()*1000

    diff := now-mseconds
    if diff<0 || diff>s.maxMSeconds {
        return //cowardly refuses
    }

    slot := mseconds - (mseconds % int64(s.msecondsPerPoint))

    if s.data[0].MSeconds == 0 { // empty 
        s.data[0].MSeconds = slot
        s.data[0].Value = value
        log.Printf("%d,%d",0,slot)
    } else {
        timeDistance := slot - s.data[0].MSeconds
        slotDistance := timeDistance / int64(s.msecondsPerPoint)
        index := slotDistance % int64(s.maxPoints)
        s.data[index].MSeconds = slot
        s.data[index].Value = value
        log.Printf("%d,%d",index,slot)
    }
}

func (s *SerieTemporal) Fetch(from int64, until int64) []Slot {
    now := time.Now().Unix()*1000
    //limiteInferior := now - s.maxSeconds

    log.Printf(":%d,%d,%d",from, until, now)
    if until <= from { // empty 
        return nil
    } 

    fromInterval := from - (from % int64(s.msecondsPerPoint)) + int64(s.msecondsPerPoint) //alineados
    untilInterval := until - 1 - (until % int64(s.msecondsPerPoint)) //alineados

    result := make([]Slot, (untilInterval - fromInterval) / int64(s.msecondsPerPoint) + 1)
    log.Printf("%d,%d,%d",fromInterval, untilInterval, len(result))
    
    if s.data[0].MSeconds == 0 { // empty 
        return result
    } 

    timeDistance := fromInterval - s.data[0].MSeconds
    slotDistanceFrom := timeDistance / int64(s.msecondsPerPoint)
    slot := slotDistanceFrom % int64(s.maxPoints)

    if slot < 0 {
        slot = slot + int64(s.maxPoints)
    }

    i:=0
    for interval:=fromInterval; interval <= untilInterval;  {
        //log.Printf("%d,%d",slot,i)
        if s.data[slot].MSeconds == interval {
            result[i].MSeconds = interval
            result[i].Value = s.data[slot].Value
        }
        interval = interval + int64(s.msecondsPerPoint)
        slot= (slot+1) % int64(s.maxPoints)
        i = i + 1
    }
    return result
}

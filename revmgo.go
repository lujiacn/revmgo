package revmgo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/revel/revel"
)

type MgoController struct {
	*revel.Controller
	MgoSession *mgo.Session
}

//package variables
var (
	MgoSession     *mgo.Session
	MgoDBName      string
	Dial           string
	ObjectIdBinder = revel.Binder{
		// Make a ObjectId from a request containing it in string format.
		Bind: revel.ValueBinder(func(val string, typ reflect.Type) reflect.Value {
			if len(val) == 0 {
				return reflect.Zero(typ)

			}
			if bson.IsObjectIdHex(val) {
				objId := bson.ObjectIdHex(val)
				return reflect.ValueOf(objId)

			} else {
				revel.AppLog.Errorf("ObjectIdBinder.Bind - invalid ObjectId!")
				return reflect.Zero(typ)

			}

		}),
		// Turns ObjectId back to hexString for reverse routing
		Unbind: func(output map[string]string, name string, val interface{}) {
			var hexStr string
			hexStr = fmt.Sprintf("%s", val.(bson.ObjectId).Hex())
			// not sure if this is too carefull but i wouldn't want invalid ObjectIds in my App
			if bson.IsObjectIdHex(hexStr) {
				output[name] = hexStr

			} else {
				revel.AppLog.Errorf("ObjectIdBinder.Unbind - invalid ObjectId!")
				output[name] = ""

			}

		},
	}
)

//MgoDBConnect initiate DB connection and set global variables
func MgoDBConnect() {
	var err error
	var found bool

	Dial = revel.Config.StringDefault("mongodb.dial", "localhost")
	if MgoDBName, found = revel.Config.String("mongodb.name"); !found {
		urls := strings.Split(Dial, "/")
		if len(urls) <= 1 {
			panic("MongoDB name not defined")
		}
		MgoDBName = urls[len(urls)-1]
	}

	MgoSession, err = mgo.Dial(Dial)

	if err != nil {
		panic("Cannot connect to database")
	}

	if MgoSession == nil {
		MgoSession, err = mgo.Dial(Dial)
		if err != nil {
			panic("Cannot connect to database")
		}
	}
}

//AppMgoInit initiate mongo db from app start
func AppMgoInit() {
	MgoDBConnect()

	objId := bson.NewObjectId()
	revel.TypeBinders[reflect.TypeOf(objId)] = ObjectIdBinder
}

//ControllerInit for revel controller
func ControllerInit() {
	revel.InterceptMethod((*MgoController).Begin, revel.BEFORE)
	revel.InterceptMethod((*MgoController).End, revel.FINALLY)
}

func (c *MgoController) Begin() revel.Result {
	if MgoSession == nil {
		MgoDBConnect()
	}
	//use clone
	c.MgoSession = MgoSession.Clone()
	return nil
}

func (c *MgoController) End() revel.Result {
	if c.MgoSession != nil {
		c.MgoSession.Close()
	}
	return nil
}

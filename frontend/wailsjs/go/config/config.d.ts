// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {config} from '../models';

export function AddFilterList(arg1:config.filterList):Promise<void>;

export function GetFilterLists():Promise<Array<config.filterList>>;

export function GetPort():Promise<number>;

export function RemoveFilterList(arg1:string):Promise<void>;

export function Save():Promise<void>;

export function SetPort(arg1:number):Promise<void>;

export function ToggleFilterList(arg1:string,arg2:boolean):Promise<void>;

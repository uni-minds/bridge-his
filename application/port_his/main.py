#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# @File  : File.py
# @Author: Ray
# @Date  : 2021/7/13
# @Desc  :
import argparse
import hashlib
import re
import sqlite3
import json

import numpy as np
import pandas as pd

class Source():
    def __init__(self, path, version):
        '''

        :param path: excel路径
        :param version: 版本名，不同的版本就是不同的表明
        '''
        self.columns = []

        self.num_col = 0
        self.num_row = 0

        self.path = path
        self.version = version
        self.get_file()

        self.idx = 0

    def get_file(self, ):
        df = pd.read_csv(self.path)
        self.df = np.array(df.dropna(how='all'))
        p1 = re.compile(r'[[].*[]]', re.S)
        for str in df.columns:
            ncol = re.sub(p1, '', str).replace(' ', '')  # 所有的列名去除中括号
            if ncol in self.columns:  # 重名后加2
                ncol += '2'
            self.columns.append(ncol)
        self.columns.insert(0, 'md5')  # 插入md5列

        self.num_row, self.num_col = self.df.shape

    def next(self, idx):
        if idx <= self.num_row:
            _row = self.df[idx]
            row = []
            for i in _row:
                if i != i:
                    i = 'nan'
                elif type(i) == np.float or type(i) == np.float64:
                    if int(i) == i:
                        i = int(i)
                row.append(str(i).replace('\'', '').replace('\"', ''))
            o1 = row[0]
            nm = row[1]
            o2 = row[0]
            uid = row[3]
            phone = row[4]
            md5hash = hashlib.md5((o1 + o2 + nm + uid + phone).encode('utf-8'))
            md5 = md5hash.hexdigest()
            row.insert(0, md5)
            ret = dict(zip(self.columns, row))
            return ret


class Sqldata():
    def __init__(self, dbname, version, columns=None):
        self.conn = sqlite3.connect(dbname)  # 数据库链接
        self.version = version
        if columns is not None:  # 指定列名即创建
            self._createdb(columns)
            self.cloumns = columns
        else:
            self._getcloumns()  # 通过查询获取当前表的列

    def _getcloumns(self):
        cur = self.conn.cursor()
        cur.execute("SELECT * FROM {};".format(self.version))
        columns_tuple = cur.description

        self.cloumns = [field_tuple[0] for field_tuple in columns_tuple]  # 获取到列名

    def _createdb(self, columns):
        query = 'CREATE TABLE IF NOT EXISTS {} ('.format(self.version)

        for k in columns:
            query += "\"{}\"".format(k) + ' ' + 'CHAR(50),'

        query += 'PRIMARY KEY( "md5" )'
        query += ');'
        cur = self.conn.cursor()

        cur.execute(query)
        self.conn.commit()

    def insert(self, row):
        query = "INSERT INTO {}".format(self.version)
        query += " ({}) VALUES ({});"
        keys = ''
        vals = ''
        for k, v in row.items():
            keys += "\"{}\",".format(k)
            vals += "\"{}\",".format(str(v))

        keys = keys[:-1]
        vals = vals[:-1].replace(' ', '')
        query = query.format(keys, vals)
        cur = self.conn.cursor()
        try:
            cur.execute(query)
        except sqlite3.IntegrityError:
            print(vals[0:60], '---插入主键（md5）冲突，数据库中已存在相关信息。')
        except sqlite3.OperationalError:
            print(query)
        self.conn.commit()

    def delete(self, ):
        pass

    def modify(self):
        pass

    def search(self, cond):
        if cond == {}:
            query = "SELECT * FROM {}".format(self.version)
        else:
            query = "SELECT * FROM {} WHERE ".format(self.version)
            # conds = "\"{}\" = \"{}\""
            conds = '''"{}" like "%{}%"'''  #sql模糊
            for k, v in cond.items():
                cd = conds.format(k, v) + " AND "
                query += cd
            query = query[0:-5]
            query += ';'
        cur = self.conn.cursor()
        # print(query)
        cur.execute(query)
        ret = []
        results = cur.fetchall()
        self.conn.commit()
        for r in results:
            ret.append(dict(zip(self.cloumns, list(r))))
        result = json.dumps(ret)
        return result


if __name__ == '__main__':

    parser = argparse.ArgumentParser(formatter_class=argparse.ArgumentDefaultsHelpFormatter)

    parser.add_argument('--is_import', action='store_true', help='Import excel.')

    parser.add_argument('--db_path', type=str, default='./db_his.sqlite', help='The path of database.')
    parser.add_argument('--csv_path', type=str, default='./source.csv', help='The path of csv.')
    parser.add_argument('--version', type=str, default='version2', help='The version of excel.')
    parser.add_argument('--uid', type=str, default='3135466', help='产妇入院登记号号码')

    opt = parser.parse_args()

    # sor = Source(opt.excel_path, version=opt.version)

    if opt.is_import:
        sor = Source(opt.excel_path, version=opt.version)
        db = Sqldata(opt.db_path, sor.version, sor.columns)
        for i in range(sor.num_row):
            db.insert(sor.next(i))
    else:
#         print("args:", opt.db_path, opt.version)
        db = Sqldata(opt.db_path, opt.version)

        results = db.search({'产妇入院登记号号码': '{}'.format(opt.uid)})
        print(results)

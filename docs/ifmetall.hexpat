#include <std/mem.pat>
#include <std/string.pat>

s128 count;

struct Header {
  char line_type[2];
  char union_no[2];
  char location[4];
  char company_code[10];
  char company_name[24];
  char acc_type[1];
  char acc_month[2];
  char acc_year[2];
  char date[6];
  char filler[15];
};


struct Detail {
  char a;
  char b;
  char union_no[2];
  char location[4];
  char id[10];
  char name[24];
  char fee[6];
  char check[6];
  char paycode[2];
  char filler[10];
  padding[2];
//  if (b == 0x33)
//    break;
};

struct Footer {
  char line_type[2];
  char union_no[2];
  char location[4];
  char company_code[10];
  char company_name[24];
  char entry_count[6];
  char sum_fee[9];
  char sum_check[9];
  padding[2];
};

Header header @ 0x00;
Footer footer @ std::mem::find_sequence(0, 0x53, 0x33);

//count = std::string::parse_int("00043", 10);

Detail details[std::string::parse_int(footer.entry_count, 10)] @ std::mem::find_sequence(0, 0x53, 0x32);




